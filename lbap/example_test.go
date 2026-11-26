package lbap_test

import (
	"fmt"

	"github.com/ryanjoneil/ap/lbap"
)

func Example() {
	a, err := lbap.New([][]int{
		{10, 15, 12},
		{51, 75, 23},
		{11, 91, 10},
	})
	if err != nil {
		panic(err)
	}
	p, err := a.Assign()
	if err != nil {
		panic(err)
	}
	fmt.Println(p)
	fmt.Println(a.Cost())
	// Output:
	// [1 2 0]
	// 23
}
