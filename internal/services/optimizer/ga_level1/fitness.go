package ga_level1

import (
	"math"
	"noytech-ga-optimizer/api/proto"
	"noytech-ga-optimizer/internal/models"
	"noytech-ga-optimizer/internal/services/optimizer/logic"
)

func CalculateFitness(
	ind *Individual,
	allTerminals []models.Terminal,
	shipments []models.Shipment,
	interCityRates []models.InterCityRate,
	intraCityRates []models.IntraCityRate,
	distances map[string]map[string]int,
) error {
	// 1. Активные терминалы
	activeTerminals := make([]models.Terminal, 0)
	for i, active := range ind.TerminalMask {
		if active {
			activeTerminals = append(activeTerminals, allTerminals[i])
		}
	}

	if len(activeTerminals) == 0 {
		ind.Cost = CostBreakdown{TotalCost: 1e12}
		ind.Fitness = 1e12
		return nil
	}

	// 2. Распределение грузов по ближайшему терминалу
	terminalShipments := make(map[string][]models.Shipment)
	penalty := 0.0
	for _, s := range shipments {
		bestCity := ""
		minDist := math.MaxInt
		for _, t := range activeTerminals {
			if distMap, ok := distances[t.City]; ok {
				if d, ok2 := distMap[s.DestinationCity]; ok2 && d < minDist {
					minDist = d
					bestCity = t.City
				}
			}
		}
		if bestCity == "" {
			penalty += 1e9
			continue
		}
		terminalShipments[bestCity] = append(terminalShipments[bestCity], s)
	}

	// 3. Last-mile cost
	lastMileCost := 0.0
	for city, sList := range terminalShipments {
		cost, err := logic.CalculateLastMileCostForTerminal(
			sList, city, interCityRates, intraCityRates, distances[city])
		if err != nil {
			return err
		}
		lastMileCost += cost
	}

	// 4. Linehaul cost
	linehaulCost := 0.0
	for _, t := range activeTerminals {
		cost, err := logic.CalculateLinehaulCost(t, activeTerminals, interCityRates)
		if err != nil {
			return err
		}
		linehaulCost += cost
	}

	// 5. Штрафы и подбор ТС
	routes := make([]RouteWithShipments, 0)
	for city, sList := range terminalShipments {
		totalWeightTons := 0.0
		totalVolumeM3 := 0.0
		ids := make([]string, 0)
		for _, s := range sList {
			totalWeightTons += s.WeightKg / 1000.0
			totalVolumeM3 += s.VolumeM3
			ids = append(ids, s.ID)
		}

		var chosenType proto.TransportType
		var capTons, capM3 float64
		found := false
		for _, tr := range logic.AvailableTransports {
			if totalWeightTons <= tr.CapTons && totalVolumeM3 <= tr.CapM3 {
				chosenType = tr.Type
				capTons = tr.CapTons
				capM3 = tr.CapM3
				found = true
				break
			}
		}

		if !found {
			last := logic.AvailableTransports[len(logic.AvailableTransports)-1]
			chosenType = last.Type
			capTons = last.CapTons
			capM3 = last.CapM3
			penalty += 50000
		}

		utilWeight := totalWeightTons / capTons
		utilVolume := totalVolumeM3 / capM3
		utilization := math.Max(utilWeight, utilVolume)

		if utilization < 0.6 {
			penalty += 10000 * (0.6 - utilization)
		}

		routes = append(routes, RouteWithShipments{
			FromCity:      "Москва",
			ToTerminal:    city,
			ShipmentIDs:   ids,
			Cost:          0,
			TransportUsed: chosenType,
		})
	}

	totalCost := linehaulCost + lastMileCost + penalty
	ind.Cost = CostBreakdown{
		LinehaulCost: linehaulCost,
		LastMileCost: lastMileCost,
		PenaltyCost:  penalty,
		TotalCost:    totalCost,
	}
	ind.Fitness = totalCost
	ind.ActiveTerminals = make([]string, len(activeTerminals))
	for i, t := range activeTerminals {
		ind.ActiveTerminals[i] = t.City
	}
	ind.Routes = routes
	return nil
}
