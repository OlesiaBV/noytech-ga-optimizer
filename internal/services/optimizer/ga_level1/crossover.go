package ga_level1

import (
	"math/rand"
	"noytech-ga-optimizer/api/proto"
)

func Crossover(p1, p2 *Individual, method proto.CrossoverType) (*Individual, *Individual) {
	var mask1, mask2 []bool
	switch method {
	case proto.CrossoverType_CROSSOVER_UNIFORM:
		mask1, mask2 = uniformCrossover(p1.TerminalMask, p2.TerminalMask)
	case proto.CrossoverType_CROSSOVER_SINGLE_POINT:
		mask1, mask2 = singlePointCrossover(p1.TerminalMask, p2.TerminalMask)
	case proto.CrossoverType_CROSSOVER_TWO_POINT:
		mask1, mask2 = twoPointCrossover(p1.TerminalMask, p2.TerminalMask)
	default:
		panic("unsupported crossover type")
	}
	return &Individual{TerminalMask: mask1}, &Individual{TerminalMask: mask2}
}

func uniformCrossover(a, b []bool) ([]bool, []bool) {
	c1, c2 := make([]bool, len(a)), make([]bool, len(a))
	for i := range a {
		if rand.Float64() < 0.5 {
			c1[i], c2[i] = a[i], b[i]
		} else {
			c1[i], c2[i] = b[i], a[i]
		}
	}
	return c1, c2
}

func singlePointCrossover(a, b []bool) ([]bool, []bool) {
	if len(a) <= 1 {
		return a, b
	}
	i := rand.Intn(len(a)-1) + 1
	c1 := append(append([]bool{}, a[:i]...), b[i:]...)
	c2 := append(append([]bool{}, b[:i]...), a[i:]...)
	return c1, c2
}

func twoPointCrossover(a, b []bool) ([]bool, []bool) {
	n := len(a)
	if n <= 2 {
		return a, b
	}
	i, j := rand.Intn(n), rand.Intn(n)
	if i > j {
		i, j = j, i
	}
	c1 := append(append(append([]bool{}, a[:i]...), b[i:j]...), a[j:]...)
	c2 := append(append(append([]bool{}, b[:i]...), a[i:j]...), b[j:]...)
	return c1, c2
}
