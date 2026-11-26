package mcm_test

import (
	"fmt"

	"github.com/ryanjoneil/ap/mcm"
)

func Example() {
	m, err := mcm.New([][]bool{
		{true, true, false},
		{true, false, false},
		{false, true, true},
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(m.Match())
	fmt.Println(m.Size())
	// Output:
	// [1 0 2]
	// 3
}
