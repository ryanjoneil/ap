package ap_test

import (
	"fmt"

	"github.com/ryanjoneil/ap"
)

func ExampleMatrix() {
	m := ap.Matrix{
		{false, true, false, false, false, false, false},
		{true, false, false, false, false, false, false},
		{false, false, true, false, false, false, false},
		{false, false, false, false, false, false, true},
		{false, false, false, false, false, true, false},
		{false, false, false, true, false, false, false},
		{false, false, false, false, true, false, false},
	}
	fmt.Println(m)
	fmt.Println(m.Cycles())
	fmt.Println(m.Inverse())
	fmt.Println(m.Permutation())

	// Output:
	// - X - - - - -
	// X - - - - - -
	// - - X - - - -
	// - - - - - - X
	// - - - - - X -
	// - - - X - - -
	// - - - - X - -
	// [[0 1] [2] [3 6 4 5]]
	// - X - - - - -
	// X - - - - - -
	// - - X - - - -
	// - - - - - X -
	// - - - - - - X
	// - - - - X - -
	// - - - X - - -
	// [1 0 2 6 5 3 4]
}
