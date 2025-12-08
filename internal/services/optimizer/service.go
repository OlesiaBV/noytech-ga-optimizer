package optimizer

import (
	"context"
	"log/slog"

	"noytech-ga-optimizer/api/proto"
	"noytech-ga-optimizer/internal/models"
	"noytech-ga-optimizer/internal/services/optimizer/ga_level1"
	"noytech-ga-optimizer/internal/services/optimizer/ga_level2"
	"noytech-ga-optimizer/internal/services/optimizer/logic"
	storage "noytech-ga-optimizer/internal/storages"
	"noytech-ga-optimizer/pkg/errors"
)

type Service struct {
	storage storage.Storage
	logger  *slog.Logger
}

func New(s storage.Storage, l *slog.Logger) *Service {
	return &Service{
		storage: s,
		logger:  l,
	}
}

func (s *Service) Optimize(ctx context.Context, req *proto.OptimizeRequest) (*proto.OptimizationResult, error) {
	logger := s.logger.With(
		slog.String("method", "Optimize"),
		slog.String("direction", req.Direction),
		slog.Any("delivery_days", req.DeliveryDays),
	)

	logger.Info("Starting optimization request")

	// 1. Загрузка всех данных из БД
	shipments, err := s.storage.GetAllShipments(ctx)
	if err != nil {
		logger.Error("Failed to load shipments", "error", err)
		return nil, errors.NewErrOptimizationFailed("failed to load shipments: %v", err)
	}

	terminals, err := s.storage.GetAllTerminals(ctx)
	if err != nil {
		logger.Error("Failed to load terminals", "error", err)
		return nil, errors.NewErrOptimizationFailed("failed to load terminals: %v", err)
	}

	distances, err := s.storage.GetAllDistances(ctx)
	if err != nil {
		logger.Error("Failed to load distances", "error", err)
		return nil, errors.NewErrOptimizationFailed("failed to load distances: %v", err)
	}

	interCityRates, err := s.storage.GetAllInterCityRates(ctx)
	if err != nil {
		logger.Error("Failed to load inter-city rates", "error", err)
		return nil, errors.NewErrOptimizationFailed("failed to load inter-city rates: %v", err)
	}

	intraCityRates, err := s.storage.GetAllIntraCityRates(ctx)
	if err != nil {
		logger.Error("Failed to load intra-city rates", "error", err)
		return nil, errors.NewErrOptimizationFailed("failed to load intra-city rates: %v", err)
	}

	// 2. Фильтрация терминалов по направлению (если указано)
	filteredTerminals := terminals
	if req.Direction != "" {
		filteredTerminals = make([]models.Terminal, 0)
		for _, t := range terminals {
			if t.Direction == req.Direction {
				filteredTerminals = append(filteredTerminals, t)
			}
		}
		if len(filteredTerminals) == 0 {
			return nil, errors.NewErrOptimizationFailed("no terminals found for direction: %s", req.Direction)
		}
	}

	// 3. Преобразуем distances
	distancesMap := make(map[string]map[string]int)
	for _, d := range distances {
		if distancesMap[d.FromCity] == nil {
			distancesMap[d.FromCity] = make(map[string]int)
		}
		distancesMap[d.FromCity][d.ToCity] = d.Km
	}

	// 4. Группируем грузы по дням отгрузки
	groupedShipments, err := logic.GroupShipmentsByDeliveryDay(shipments, req.DeliveryDays)
	if err != nil {
		logger.Error("Failed to group shipments", "error", err)
		return nil, errors.NewErrOptimizationFailed("grouping failed: %v", err)
	}

	// 5. Будем хранить лучший результат среди дней
	var bestResult *proto.OptimizationResult
	var bestCost float64 = 1e18

	// Запуск оптимизации для каждого дня отгрузки
	for deliveryDay, dayShipments := range groupedShipments {
		logger.Info("Optimizing for delivery day", "day", deliveryDay, "shipment_count", len(dayShipments))

		// Уровень 1: выбор терминалов
		level1Result, err := ga_level1.RunGA(
			req.GaSettingsLevel_1,
			filteredTerminals,
			dayShipments,
			interCityRates,
			intraCityRates,
			distancesMap,
		)
		if err != nil {
			logger.Error("Level 1 GA failed", "day", deliveryDay, "error", err)
			return nil, errors.NewErrOptimizationFailed("level 1 GA failed: %v", err)
		}

		var activeTerminals []models.Terminal
		for _, city := range level1Result.ActiveTerminals {
			for _, t := range filteredTerminals {
				if t.City == city {
					activeTerminals = append(activeTerminals, t)
					break
				}
			}
		}

		// Уровень 2: расчёт стоимости для фиксированного набора терминалов
		level2Result, err := ga_level2.RunGALevel2(
			activeTerminals,
			dayShipments,
			interCityRates,
			intraCityRates,
			distancesMap,
		)
		if err != nil {
			logger.Error("Level 2 GA failed", "day", deliveryDay, "error", err)
			return nil, errors.NewErrOptimizationFailed("level 2 GA failed: %v", err)
		}

		protoResult := s.convertToProto(level2Result, 0)

		if level2Result.Fitness < bestCost {
			bestCost = level2Result.Fitness
			bestResult = protoResult
		}
	}

	if bestResult == nil {
		return nil, errors.NewErrOptimizationFailed("no valid result produced")
	}

	logger.Info("Optimization completed successfully", "best_total_cost", bestCost)
	return bestResult, nil
}

func (s *Service) convertToProto(level2 *ga_level2.Individual, generation int32) *proto.OptimizationResult {
	routes := make([]*proto.Route, len(level2.Routes))
	for i, r := range level2.Routes {
		routes[i] = &proto.Route{
			FromCity:      r.FromCity,
			ToTerminal:    r.ToTerminal,
			ShipmentIds:   r.ShipmentIDs,
			Cost:          r.Cost,
			TransportUsed: r.TransportUsed,
		}
	}

	return &proto.OptimizationResult{
		Routes:          routes,
		Cost:            convertCost(level2.Cost),
		ActiveTerminals: level2.ActiveTerminals,
		Generation:      generation,
		FitnessScore:    level2.Fitness,
	}
}

func convertCost(cost ga_level2.CostBreakdown) *proto.CostBreakdown {
	return &proto.CostBreakdown{
		LinehaulCost: cost.LinehaulCost,
		LastMileCost: cost.LastMileCost,
		PenaltyCost:  cost.PenaltyCost,
		TotalCost:    cost.TotalCost,
	}
}
