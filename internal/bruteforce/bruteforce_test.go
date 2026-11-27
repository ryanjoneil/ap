package bruteforce

import (
	"slices"
	"testing"
)

var c = [][]int64{
	{10, 15, 12},
	{51, 75, 23},
	{11, 91, 10},
}

func TestMinSum(t *testing.T) {
	if best, ok := MinSum(c, nil); best != 49 || !ok {
		t.Errorf("MinSum = (%v, %v); want (49, true)", best, ok)
	}

	mask := [][]bool{{false, false, false}, {true, true, true}, {true, true, true}}
	if _, ok := MinSum(c, mask); ok {
		t.Error("MinSum with row 0 forbidden: want infeasible")
	}
}

func TestMinMax(t *testing.T) {
	if best, ok := MinMax(c, nil); best != 23 || !ok {
		t.Errorf("MinMax = (%v, %v); want (23, true)", best, ok)
	}
}

func TestMaxMatching(t *testing.T) {
	if m := MaxMatching([][]bool{{true, false, true}, {true, false, false}}); m != 2 {
		t.Errorf("MaxMatching = %v; want 2", m)
	}
	if m := MaxMatching([][]bool{{true}, {true}}); m != 1 {
		t.Errorf("MaxMatching = %v; want 1", m)
	}
}

func TestAllSums(t *testing.T) {
	c := [][]int64{{1, 2}, {3, 4}}
	if got := AllSums(c, nil); !slices.Equal(got, []int64{5, 5}) {
		t.Errorf("AllSums = %v; want [5 5]", got)
	}
	if got := AllSums(c, [][]bool{{false, true}, {true, true}}); !slices.Equal(got, []int64{5}) {
		t.Errorf("AllSums with (0,0) forbidden = %v; want [5]", got)
	}
	if got := AllSums(c, [][]bool{{false, false}, {true, true}}); got != nil {
		t.Errorf("AllSums with row 0 forbidden = %v; want nil", got)
	}
}
