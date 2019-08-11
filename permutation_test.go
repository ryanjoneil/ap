package ap_test

import (
	"fmt"

	"github.com/ryanjoneil/ap"
)

func ExamplePermutation() {
	p := ap.Permutation{1, 0, 2, 6, 5, 3, 4}
	fmt.Println(p)
	fmt.Println(p.Cycles())
	fmt.Println(p.Inverse())
	fmt.Println(p.Matrix())

	// Output:
	// [1 0 2 6 5 3 4]
	// [[0 1] [2] [3 6 4 5]]
	// [1 0 2 5 6 4 3]
	// [[0 1 0 0 0 0 0] [1 0 0 0 0 0 0] [0 0 1 0 0 0 0] [0 0 0 0 0 0 1] [0 0 0 0 0 1 0] [0 0 0 1 0 0 0] [0 0 0 0 1 0 0]]
}
