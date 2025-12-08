package ga_level1

import (
	"noytech-ga-optimizer/api/proto"
	"noytech-ga-optimizer/internal/models"
)

func RunGA(
	settings *proto.GASettings,
	terminals []models.Terminal,
	shipments []models.Shipment,
	interCityRates []models.InterCityRate,
	intraCityRates []models.IntraCityRate,
	distances map[string]map[string]int,
) (*Individual, error) {
	pop := NewRandomPopulation(int(settings.NumIndividuals), terminals)
	if err := pop.Evaluate(shipments, interCityRates, intraCityRates, distances); err != nil {
		return nil, err
	}

	best := pop.GetBest()
	noImprove := 0

	for gen := 0; gen < int(settings.NumGenerations); gen++ {
		currentBest := pop.GetBest()
		if currentBest.Fitness < best.Fitness {
			best = currentBest
			noImprove = 0
		} else {
			noImprove++
		}

		if noImprove >= int(settings.StoppingCriterion) {
			break
		}

		parents := SelectParents(pop.Individuals, len(pop.Individuals), settings.SelectionType)
		newPop := make([]*Individual, 0, len(pop.Individuals))

		for i := 0; i < len(parents); i += 2 {
			p1 := parents[i]
			p2 := parents[(i+1)%len(parents)]
			child1, child2 := Crossover(p1, p2, settings.CrossoverType)
			Mutate(child1, 0.1, settings.MutationType)
			Mutate(child2, 0.1, settings.MutationType)
			newPop = append(newPop, child1, child2)
		}

		if len(newPop) > len(pop.Individuals) {
			newPop = newPop[:len(pop.Individuals)]
		}

		pop.Individuals = newPop
		if err := pop.Evaluate(shipments, interCityRates, intraCityRates, distances); err != nil {
			return nil, err
		}
	}

	return best, nil
}
