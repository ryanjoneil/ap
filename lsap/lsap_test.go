package lsap_test

import (
	"fmt"

	"github.com/ryanjoneil/ap/lsap"
)

func Example() {
	a, err := lsap.New([][]int{
		{10, 15, 12},
		{51, 75, 23},
		{11, 91, 10},
	})
	if err != nil {
		panic(err)
	}

	// Solve the LSAP and get information about the optimal solution.
	p, err := a.Assign()
	if err != nil {
		panic(err)
	}
	fmt.Println(p)
	fmt.Println(a.Cost())
	fmt.Println(a.DualPrices())
	rc, _ := a.ReducedCost(1, 1)
	fmt.Println(rc)

	// Forbid an edge of the optimal solution and re-solve.
	a.Fix(0, 1, false)
	p, err = a.Assign()
	if err != nil {
		panic(err)
	}
	fmt.Println(p)
	fmt.Println(a.Cost())
	// Output:
	// [1 2 0]
	// 49
	// {[0 14 1] [10 15 9]}
	// 46
	// [0 1 2]
	// 95
}
