package lsap_test

import (
	"reflect"
	"testing"

	"github.com/ryanjoneil/ap"
	"github.com/ryanjoneil/ap/lsap"
)

func TestAssigner(t *testing.T) {
	a := lsap.New([][]int64{
		{25, 40, 9, 15},
		{12, 33, 22, 31},
		{18, 37, 18, 30},
		{32, 11, 12, 26},
	})

	// Optimize the LSAP and check values.
	checkAssigner(t, a, []int{3, 0, 2, 1}) // Side effect: optimizes problem.
	checkCoster(t, a, 56)
	checkDualPricer(t, a, ap.DualPrices[int64]{
		U: []int64{0, 3, 9, 0},
		V: []int64{9, 11, 9, 15},
	})
	checkReducedCoster(t, a, [][]int64{
		{16, 29, 0, 0},
		{0, 19, 10, 13},
		{0, 17, 0, 6},
		{23, 0, 3, 11},
	})

	// Remove edges, re-optimize, and check.
	a.Remove(2, 2)
	a.Remove(1, 0)
	a.Remove(0, 3)
	a.Remove(3, 1)
	a.Remove(2, 0)
	a.Remove(3, 3)

	checkAssigner(t, a, []int{0, 1, 3, 2}) // Side effect: optimizes problem.
	checkCoster(t, a, 100)
	checkDualPricer(t, a, ap.DualPrices[int64]{
		U: []int64{16, 22, 21, 19},
		V: []int64{9, 11, -7, 9},
	})
	checkReducedCoster(t, a, [][]int64{
		{0, 13, 0, 2305843009213693926},
		{2305843009213693920, 0, 7, 0},
		{2305843009213693921, 5, 2305843009213693937, 0},
		{4, 2305843009213693921, 0, 2305843009213693923},
	})
}

func checkAssigner(t *testing.T, a ap.Assigner, want ap.Permutation) {
	t.Helper()

	if v := a.Assign(); !reflect.DeepEqual(want, v) {
		t.Errorf("want %v; got %v", want, v)
	}
}

func checkCoster[T ap.Integer](t *testing.T, a ap.Coster[T], want T) {
	t.Helper()

	if v := a.Cost(); want != v {
		t.Errorf("want %v; got %v", want, v)
	}
}

func checkDualPricer[T ap.Integer](
	t *testing.T,
	a ap.DualPricer[T],
	want ap.DualPrices[T],
) {
	t.Helper()

	if v := a.DualPrices(); !reflect.DeepEqual(want, v) {
		t.Errorf("want %v; got %v", want, v)
	}
}

func checkReducedCoster[T ap.Integer](
	t *testing.T,
	a ap.ReducedCoster[T],
	want [][]T,
) {
	t.Helper()

	for u, row := range want {
		for v, rc := range row {
			if vrc := a.ReducedCost(u, v); vrc != rc {
				t.Errorf("want rc[%v][%v] = %v; got %v", u, v, rc, vrc)
			}
		}
	}
}
