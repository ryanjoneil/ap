package ap_test

import (
	"fmt"

	"github.com/ryanjoneil/ap"
)

func ExamplePermutation_Cycles() {
	p := ap.Permutation{1, 0, 2, 6, 5, 3, 4}
	fmt.Println(p.Cycles())

	// Output:
	// [[0 1] [2] [3 6 4 5]]
}

func ExamplePermutation_Inverse() {
	p := ap.Permutation{1, 0, 2, 6, 5, 3, 4}
	fmt.Println(p.Inverse())
	// Output:
	// [1 0 2 5 6 4 3]
}

func ExamplePermutation_Matrix() {
	p := ap.Permutation{1, 0, 2, 6, 5, 3, 4}
	for _, row := range p.Matrix() {
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
