package lsap_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/ryanjoneil/ap"
	"github.com/ryanjoneil/ap/lsap"
)

func ExampleLSAP() {
	a := lsap.New([][]int64{
		{10, 15, 12},
		{51, 75, 23},
		{11, 91, 10},
	})

	// Solve the LSAP and get information about the optimal solution.
	fmt.Println(a.Assign())
	fmt.Println(a.Cost())
	fmt.Println(a.DualPrices())
	fmt.Println(a.ReducedCost(1, 1))

	// Remove an edge from the optimal solution and re-solve.
	a.Remove(0, 1)
	fmt.Println(a.Assign())
	fmt.Println(a.Cost())

	// Output:
	// [1 2 0]
	// 49
	// {[0 14 1] [10 15 9]}
	// 46
	// [0 1 2]
	// 95
}

func TestLSAP(t *testing.T) {
	a := lsap.New([][]int64{
		{25, 40, 9, 15},
		{12, 33, 22, 31},
		{18, 37, 18, 30},
		{32, 11, 12, 26},
	})

	// Optimize the LSAP and check values.
	checkAssigner(t, a, []int{3, 0, 2, 1}) // Side effect: optimizes problem.
	checkInt64Coster(t, a, 56)
	checkInt64DualPricer(t, a, ap.Int64DualPrices{
		U: []int64{0, 3, 9, 0},
		V: []int64{9, 11, 9, 15},
	})
	checkInt64ReducedCoster(t, a, [][]int64{
		{16, 29, 0, 0},
		{0, 19, 10, 13},
		{0, 17, 0, 6},
		{23, 0, 3, 11},
	})

	// Remove edges, re-optimize, and check.
	a.M = 1000
	a.Remove(2, 2)
	a.Remove(1, 0)
	a.Remove(0, 3)
	a.Remove(3, 1)
	a.Remove(2, 0)
	a.Remove(3, 3)

	checkAssigner(t, a, []int{0, 1, 3, 2}) // Side effect: optimizes problem.
	checkInt64Coster(t, a, 100)
	checkInt64DualPricer(t, a, ap.Int64DualPrices{
		U: []int64{16, 22, 21, 19},
		V: []int64{9, 11, -7, 9},
	})
	checkInt64ReducedCoster(t, a, [][]int64{
		{0, 13, 0, 975},
		{969, 0, 7, 0},
		{970, 5, 986, 0},
		{4, 970, 0, 972},
	})
}

func checkAssigner(t *testing.T, a ap.Assigner, want ap.Permutation) {
	if v := a.Assign(); !reflect.DeepEqual(want, v) {
		t.Errorf("want %v; got %v", want, v)
	}
}

func checkInt64Coster(t *testing.T, a ap.Int64Coster, want int64) {
	if v := a.Cost(); want != v {
		t.Errorf("want %v; got %v", want, v)
	}
}

func checkInt64DualPricer(t *testing.T, a ap.Int64DualPricer, want ap.Int64DualPrices) {
	if v := a.DualPrices(); !reflect.DeepEqual(want, v) {
		t.Errorf("want %v; got %v", want, v)
	}
}

func checkInt64ReducedCoster(t *testing.T, a ap.Int64ReducedCoster, want [][]int64) {
	for u, row := range want {
		for v, rc := range row {
			if vrc := a.ReducedCost(u, v); vrc != rc {
				t.Errorf("want rc[%v][%v] = %v; got %v", u, v, rc, vrc)
			}
		}
	}
}
