package ap_test

import (
	"fmt"

	"github.com/ryanjoneil/ap"
)

func ExampleMatrix() {
	m := ap.Matrix{
		{0, 1, 0, 0, 0, 0, 0},
		{1, 0, 0, 0, 0, 0, 0},
		{0, 0, 1, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 1},
		{0, 0, 0, 0, 0, 1, 0},
		{0, 0, 0, 1, 0, 0, 0},
		{0, 0, 0, 0, 1, 0, 0},
	}
	fmt.Println(m)
	fmt.Println(m.Cycles())
	fmt.Println(m.Inverse())
	fmt.Println(m.Permutation())

	// Output:
	// [[0 1 0 0 0 0 0] [1 0 0 0 0 0 0] [0 0 1 0 0 0 0] [0 0 0 0 0 0 1] [0 0 0 0 0 1 0] [0 0 0 1 0 0 0] [0 0 0 0 1 0 0]]
	// [[0 1] [2] [3 6 4 5]]
	// [[0 1 0 0 0 0 0] [1 0 0 0 0 0 0] [0 0 1 0 0 0 0] [0 0 0 0 0 1 0] [0 0 0 0 0 0 1] [0 0 0 0 1 0 0] [0 0 0 1 0 0 0]]
	// [1 0 2 6 5 3 4]
}
