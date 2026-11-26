package lsap_test

import (
	"reflect"
	"testing"

	"github.com/ryanjoneil/ap"
	"github.com/ryanjoneil/ap/lsap"
)

func TestAssigner(t *testing.T) {
	a, err := lsap.New([][]int64{
		{25, 40, 9, 15},
		{12, 33, 22, 31},
		{18, 37, 18, 30},
		{32, 11, 12, 26},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Optimize the LSAP and check values.
	checkAssign(t, a, ap.Permutation{3, 0, 2, 1}) // Side effect: optimizes problem.
	checkCost(t, a, 56)
	checkDualPrices(t, a, ap.DualPrices[int64]{
		U: []int64{0, 3, 9, 0},
		V: []int64{9, 11, 9, 15},
	})
	checkReducedCosts(t, a, [][]int64{
		{16, 29, 0, 0},
		{0, 19, 10, 13},
		{0, 17, 0, 6},
		{23, 0, 3, 11},
	}, nil)

	// Remove edges, re-optimize, and check.
	a.Remove(2, 2)
	a.Remove(1, 0)
	a.Remove(0, 3)
	a.Remove(3, 1)
	a.Remove(2, 0)
	a.Remove(3, 3)

	checkAssign(t, a, ap.Permutation{0, 1, 3, 2}) // Side effect: optimizes problem.
	checkCost(t, a, 100)
	checkDualPrices(t, a, ap.DualPrices[int64]{
		U: []int64{16, 22, 21, 19},
		V: []int64{9, 11, -7, 9},
	})
	checkReducedCosts(t, a, [][]int64{
		{0, 13, 0, 0},
		{0, 0, 7, 0},
		{0, 5, 0, 0},
		{4, 0, 0, 0},
	}, [][]bool{
		{true, true, true, false},
		{false, true, true, true},
		{false, true, false, true},
		{true, false, true, false},
	})
}

func checkAssign(t *testing.T, a *lsap.LSAP[int64], want ap.Permutation) {
	t.Helper()

	p, err := a.Assign()
	if err != nil {
		t.Fatalf("Assign: %v", err)
	}
	if !reflect.DeepEqual(want, p) {
		t.Errorf("want %v; got %v", want, p)
	}
}

func checkCost(t *testing.T, a *lsap.LSAP[int64], want int64) {
	t.Helper()

	if v := a.Cost(); want != v {
		t.Errorf("want %v; got %v", want, v)
	}
}

func checkDualPrices(t *testing.T, a *lsap.LSAP[int64], want ap.DualPrices[int64]) {
	t.Helper()

	if v := a.DualPrices(); !reflect.DeepEqual(want, v) {
		t.Errorf("want %v; got %v", want, v)
	}
}

// checkReducedCosts checks every reduced cost against want. If allowed is
// non-nil, an edge with allowed[u][v] == false must report ok == false.
func checkReducedCosts(t *testing.T, a *lsap.LSAP[int64], want [][]int64, allowed [][]bool) {
	t.Helper()

	for u, row := range want {
		for v, rc := range row {
			wantOK := allowed == nil || allowed[u][v]
			got, ok := a.ReducedCost(u, v)
			switch {
			case ok != wantOK:
				t.Errorf("rc[%v][%v]: want ok = %v; got %v", u, v, wantOK, ok)
			case ok && got != rc:
				t.Errorf("want rc[%v][%v] = %v; got %v", u, v, rc, got)
			}
		}
	}
}
