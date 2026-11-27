package klsap_test

import (
	"errors"
	"fmt"

	"github.com/ryanjoneil/ap"
	"github.com/ryanjoneil/ap/klsap"
)

func Example() {
	a, err := klsap.New([][]int{
		{90, 76, 75, 70},
		{35, 85, 55, 65},
		{125, 95, 90, 105},
		{45, 110, 95, 115},
	}, 3)
	if err != nil {
		panic(err)
	}
	for {
		p, err := a.Next()
		if errors.Is(err, ap.ErrInfeasible) {
			break // No more assignments.
		}
		if err != nil {
			panic(err)
		}
		fmt.Println(p, a.Cost())
	}
	// Output:
	// [3 2 1 0] 265
	// [1 3 2 0] 276
	// [2 3 1 0] 280
}
