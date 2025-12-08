package ga_level1

import (
	"math/rand"
	"noytech-ga-optimizer/api/proto"
)

func Mutate(ind *Individual, prob float64, method proto.MutationType) {
	if rand.Float64() > prob {
		return
	}
	switch method {
	case proto.MutationType_MUTATION_INVERSION:
		inversion(ind.TerminalMask)
	case proto.MutationType_MUTATION_SWAP:
		swap(ind.TerminalMask)
	default:
		panic("unsupported mutation type")
	}
}

func inversion(mask []bool) {
	if len(mask) < 2 {
		return
	}
	i, j := rand.Intn(len(mask)), rand.Intn(len(mask))
	if i > j {
		i, j = j, i
	}
	for a, b := i, j; a < b; a, b = a+1, b-1 {
		mask[a], mask[b] = mask[b], mask[a]
	}
}

func swap(mask []bool) {
	if len(mask) < 2 {
		return
	}
	i, j := rand.Intn(len(mask)), rand.Intn(len(mask))
	mask[i], mask[j] = mask[j], mask[i]
}
