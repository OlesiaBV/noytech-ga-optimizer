package ga_level1

import (
	"math/rand"
	"noytech-ga-optimizer/api/proto"
	"sort"
)

func SelectParents(pop []*Individual, count int, method proto.SelectionType) []*Individual {
	switch method {
	case proto.SelectionType_SELECTION_TOURNAMENT:
		return tournamentSelection(pop, count)
	case proto.SelectionType_SELECTION_ROULETTE:
		return rouletteWheelSelection(pop, count)
	case proto.SelectionType_SELECTION_RANK:
		return rankSelection(pop, count)
	default:
		panic("unsupported selection type")
	}
}

func tournamentSelection(pop []*Individual, count int) []*Individual {
	parents := make([]*Individual, count)
	tSize := 3
	for i := 0; i < count; i++ {
		tour := make([]*Individual, tSize)
		for j := 0; j < tSize; j++ {
			tour[j] = pop[rand.Intn(len(pop))]
		}
		best := tour[0]
		for _, p := range tour[1:] {
			if p.Fitness < best.Fitness {
				best = p
			}
		}
		parents[i] = best
	}
	return parents
}

func rouletteWheelSelection(pop []*Individual, count int) []*Individual {
	parents := make([]*Individual, count)
	total := 0.0
	for _, p := range pop {
		total += 1.0 / (1.0 + p.Fitness)
	}
	for i := 0; i < count; i++ {
		r := rand.Float64() * total
		cum := 0.0
		for _, p := range pop {
			cum += 1.0 / (1.0 + p.Fitness)
			if cum >= r {
				parents[i] = p
				break
			}
		}
	}
	return parents
}

func rankSelection(pop []*Individual, count int) []*Individual {
	parents := make([]*Individual, count)
	sorted := make([]*Individual, len(pop))
	copy(sorted, pop)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Fitness < sorted[j].Fitness
	})
	n := len(sorted)
	rankSum := float64(n*(n+1)) / 2.0
	for i := 0; i < count; i++ {
		r := rand.Float64() * rankSum
		rank := 0.0
		for j, p := range sorted {
			rank += float64(n - j)
			if rank >= r {
				parents[i] = p
				break
			}
		}
	}
	return parents
}
