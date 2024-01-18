package lsap_test

import (
	"fmt"

	"github.com/ryanjoneil/ap/lsap"
)

func Example() {
	a := lsap.New([][]int{
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
