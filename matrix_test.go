package ap_test

import (
	"fmt"

	"github.com/ryanjoneil/ap"
)

func ExampleMatrix() {
	m := ap.Matrix{
		{false, true, false},
		{true, false, false},
		{false, false, true},
	}
	fmt.Println(m)
	fmt.Println(m.Permutation())

	// Output:
	// - X -
	// X - -
	// - - X
	// [1 0 2]
}
