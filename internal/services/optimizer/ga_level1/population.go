package ga_level1

import (
	"math/rand"
	"noytech-ga-optimizer/internal/models"
	"sort"
)

type Population struct {
	Individuals  []*Individual
	AllTerminals []models.Terminal
}

func NewRandomPopulation(size int, terminals []models.Terminal) *Population {
	pop := &Population{
		Individuals:  make([]*Individual, size),
		AllTerminals: terminals,
	}
	for i := 0; i < size; i++ {
		mask := make([]bool, len(terminals))
		for j := range mask {
			mask[j] = rand.Float32() < 0.3
		}
		pop.Individuals[i] = &Individual{TerminalMask: mask}
	}
	return pop
}

func (p *Population) Evaluate(
	shipments []models.Shipment,
	interCityRates []models.InterCityRate,
	intraCityRates []models.IntraCityRate,
	distances map[string]map[string]int,
) error {
	for _, ind := range p.Individuals {
		if err := CalculateFitness(ind, p.AllTerminals, shipments, interCityRates, intraCityRates, distances); err != nil {
			return err
		}
	}
	return nil
}

func (p *Population) GetBest() *Individual {
	best := p.Individuals[0]
	for _, ind := range p.Individuals[1:] {
		if ind.Fitness < best.Fitness {
			best = ind
		}
	}
	return best
}

func (p *Population) SortByFitness() {
	sort.Slice(p.Individuals, func(i, j int) bool {
		return p.Individuals[i].Fitness < p.Individuals[j].Fitness
	})
}
