package ap_test

import (
	"fmt"

	"github.com/ryanjoneil/ap"
)

func ExampleMatrix() {
	a := ap.Assignment{1, 3, 2, 0}
	fmt.Println(ap.ToMatrix(a))
	// Output:
	// - X - -
	// - - - X
	// - - X -
	// X - - -
}
