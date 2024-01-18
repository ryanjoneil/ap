package ap_test

import (
	"fmt"

	"github.com/ryanjoneil/ap"
)

func ExampleMatrix_Cycles() {
	m := ap.Matrix{
		[]bool{true, false, false, false, false},
		[]bool{false, false, false, false, true},
		[]bool{false, false, false, true, false},
		[]bool{false, true, false, false, false},
		[]bool{false, false, true, false, false},
	}
	fmt.Println(m.Cycles())
	// Output:
	// [[0] [1 4 2 3]]
}

func ExampleMatrix_Inverse() {
	m := ap.Matrix{
		[]bool{true, false, false, false, false},
		[]bool{false, false, false, false, true},
		[]bool{false, false, false, true, false},
		[]bool{false, true, false, false, false},
		[]bool{false, false, true, false, false},
	}
	for _, row := range m.Inverse() {
		fmt.Println(row)
	}
	// Output:
	// [true false false false false]
	// [false false false true false]
	// [false false false false true]
	// [false false true false false]
	// [false true false false false]
}

func ExampleMatrix_Permutation() {
	m := ap.Matrix{
		[]bool{true, false, false, false, false},
		[]bool{false, false, false, false, true},
		[]bool{false, false, false, true, false},
		[]bool{false, true, false, false, false},
		[]bool{false, false, true, false, false},
	}
	fmt.Println(m.Permutation())
	// Output:
	// [0 4 3 1 2]
}
