package logic

import (
	"fmt"
	"math"

	"noytech-ga-optimizer/internal/models"
)

func CalculateLinehaulCost(
	terminal models.Terminal,
	terminals []models.Terminal,
	interCityRates []models.InterCityRate,
) (float64, error) {

	if len(interCityRates) == 0 {
		return 0, fmt.Errorf("inter-city rates list is empty")
	}
	if len(terminals) == 0 {
		return 0, fmt.Errorf("terminals list is empty")
	}

	// Шаг 1: Находим min и max тариф за км
	minRate, maxRate := findMinMaxRate(interCityRates)

	// Шаг 2: Находим min и max расстояние от Москвы до терминалов
	minDest, maxDest := findMinMaxDistance(terminals)

	// Шаг 3: Получаем расстояние для текущего терминала
	distance := float64(terminal.DistanceFromMoscowKm)

	if maxDest == minDest {
		rate := (minRate + maxRate) / 2.0
		return rate * distance, nil
	}

	// Шаг 4: Применяем формулу для расчёта тарифа за км для текущего терминала
	rate := maxRate - (maxRate-minRate)*((distance-minDest)/(maxDest-minDest))

	// Шаг 5: Считаем итоговую стоимость
	cost := rate * distance

	return cost, nil
}

func findMinMaxRate(rates []models.InterCityRate) (min, max float64) {
	min = math.MaxFloat64
	max = -math.MaxFloat64

	for _, r := range rates {
		if r.RatePerKm < min {
			min = r.RatePerKm
		}
		if r.RatePerKm > max {
			max = r.RatePerKm
		}
	}

	if min == math.MaxFloat64 {
		min = 0
		max = 0
	}

	return min, max
}

func findMinMaxDistance(terminals []models.Terminal) (min, max float64) {
	min = math.MaxFloat64
	max = -math.MaxFloat64

	for _, t := range terminals {
		dist := float64(t.DistanceFromMoscowKm)
		if dist < min {
			min = dist
		}
		if dist > max {
			max = dist
		}
	}

	if min == math.MaxFloat64 {
		min = 0
		max = 0
	}

	return min, max
}
