package ap_test

import (
	"fmt"

	"github.com/ryanjoneil/ap"
)

func ExampleCycles_Inverse() {
	c := ap.Cycles{{0, 1}, {2}, {3, 6, 4, 5}}
	fmt.Println(c.Inverse())
	// Output:
	// [[0 1] [2] [3 5 4 6]]
}

func ExampleCycles_Matrix() {
	c := ap.Cycles{{0, 1}, {2}, {3, 6, 4, 5}}
	for _, row := range c.Matrix() {
		fmt.Println(row)
	}
	// Output:
	// [false true false false false false false]
	// [true false false false false false false]
	// [false false true false false false false]
	// [false false false false false false true]
	// [false false false false false true false]
	// [false false false true false false false]
	// [false false false false true false false]
}

func ExampleCycles_Permutation() {
	c := ap.Cycles{{0, 1}, {2}, {3, 6, 4, 5}}
	fmt.Println(c.Permutation())
	// Output:
	// [1 0 2 6 5 3 4]
}
