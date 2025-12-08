package ga_level2

import (
	"noytech-ga-optimizer/internal/models"
)

func RunGALevel2(
	activeTerminals []models.Terminal,
	shipments []models.Shipment,
	interCityRates []models.InterCityRate,
	intraCityRates []models.IntraCityRate,
	distances map[string]map[string]int,
) (*Individual, error) {
	if len(activeTerminals) == 0 {
		return &Individual{
			Cost:    CostBreakdown{TotalCost: 1e12},
			Fitness: 1e12,
		}, nil
	}

	return CalculateFitnessLevel2(
		activeTerminals,
		shipments,
		interCityRates,
		intraCityRates,
		distances,
	)
}
