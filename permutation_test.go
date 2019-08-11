package ap_test

import (
	"fmt"

	"github.com/ryanjoneil/ap"
)

func ExamplePermutation() {
	p := ap.Permutation{1, 3, 2, 0}
	fmt.Println(p)
	fmt.Println(p.Inverse())
	fmt.Println(p.Matrix())

	// Output:
	// [1 3 2 0]
	// [3 0 2 1]
	// - X - -
	// - - - X
	// - - X -
	// X - - -
}
