package ap_test

import (
	"fmt"

	"github.com/ryanjoneil/ap"
)

func ExampleCycles() {
	c := ap.Cycles{{0, 1}, {2}, {3, 6, 4, 5}}
	fmt.Println(c)
	fmt.Println(c.Inverse())
	fmt.Println(c.Matrix())
	fmt.Println(c.Permutation())

	// Output:
	// [[0 1] [2] [3 6 4 5]]
	// [[0 1] [2] [3 5 4 6]]
	// [[0 1 0 0 0 0 0] [1 0 0 0 0 0 0] [0 0 1 0 0 0 0] [0 0 0 0 0 0 1] [0 0 0 0 0 1 0] [0 0 0 1 0 0 0] [0 0 0 0 1 0 0]]
	// [1 0 2 6 5 3 4]
}
