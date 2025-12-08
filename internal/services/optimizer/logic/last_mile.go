package logic

import (
	"fmt"
	"sort"

	"noytech-ga-optimizer/internal/models"
)

func CalculateLastMileCostForTerminal(
	shipments []models.Shipment,
	terminalCity string,
	interCityRates []models.InterCityRate,
	intraCityRates []models.IntraCityRate,
	distances map[string]int,
) (float64, error) {

	if len(shipments) == 0 {
		return 0.0, nil
	}

	// Шаг 1: Считаем общий вес (в тоннах) и объём
	totalWeightTons := 0.0
	totalVolumeM3 := 0.0
	for _, s := range shipments {
		totalWeightTons += s.WeightKg / 1000.0
		totalVolumeM3 += s.VolumeM3
	}

	// Шаг 2: Проверяем условие из ТЗ
	if totalWeightTons > 2.5 {
		allWithin100Km := true
		for _, shipment := range shipments {
			distanceKm, exists := distances[shipment.DestinationCity]
			if !exists {
				return 0, fmt.Errorf("distance not found for route %s -> %s", terminalCity, shipment.DestinationCity)
			}
			if distanceKm > 100 {
				allWithin100Km = false
				break
			}
		}
		if allWithin100Km {
			return 7000.0, nil
		}
	}

	// Шаг 3: Подбираем один тариф для всего потока
	interCityRate, err := findInterCityRate(totalWeightTons, totalVolumeM3, interCityRates)
	if err != nil {
		return 0, fmt.Errorf("failed to find inter-city rate for total weight %.2f t and volume %.2f m3: %w", totalWeightTons, totalVolumeM3, err)
	}

	intraCityRate, err := findIntraCityRate(totalWeightTons, totalVolumeM3, intraCityRates)
	if err != nil {
		return 0, fmt.Errorf("failed to find intra-city rate for total weight %.2f t and volume %.2f m3: %w", totalWeightTons, totalVolumeM3, err)
	}

	// Шаг 4: Считаем итоговую стоимость
	var totalCost float64

	// Суммируем стоимость по каждому маршруту (терминал -> город)
	for _, shipment := range shipments {
		distanceKm, exists := distances[shipment.DestinationCity]
		if !exists {
			return 0, fmt.Errorf("distance not found for route %s -> %s", terminalCity, shipment.DestinationCity)
		}
		totalCost += interCityRate.RatePerKm * float64(distanceKm)
	}

	// Плюсуем фиксированный тариф за внутригород
	totalCost += intraCityRate.RateFixed

	return totalCost, nil
}

func findInterCityRate(weightTons, volumeM3 float64, rates []models.InterCityRate) (models.InterCityRate, error) {
	if len(rates) == 0 {
		return models.InterCityRate{}, fmt.Errorf("inter-city rates list is empty")
	}

	sorted := make([]models.InterCityRate, len(rates))
	copy(sorted, rates)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].WeightTons == sorted[j].WeightTons {
			return sorted[i].VolumeM3 < sorted[j].VolumeM3
		}
		return sorted[i].WeightTons < sorted[j].WeightTons
	})

	for _, r := range sorted {
		if weightTons <= r.WeightTons && volumeM3 <= r.VolumeM3 {
			return r, nil
		}
	}

	last := sorted[len(sorted)-1]
	return last, nil
}

func findIntraCityRate(weightTons, volumeM3 float64, rates []models.IntraCityRate) (models.IntraCityRate, error) {
	if len(rates) == 0 {
		return models.IntraCityRate{}, fmt.Errorf("intra-city rates list is empty")
	}

	sorted := make([]models.IntraCityRate, len(rates))
	copy(sorted, rates)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].WeightTons == sorted[j].WeightTons {
			return sorted[i].VolumeM3 < sorted[j].VolumeM3
		}
		return sorted[i].WeightTons < sorted[j].WeightTons
	})

	for _, r := range sorted {
		if weightTons <= r.WeightTons && volumeM3 <= r.VolumeM3 {
			return r, nil
		}
	}

	last := sorted[len(sorted)-1]
	return last, nil
}
