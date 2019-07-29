package ap_test

import (
	"fmt"

	"github.com/ryanjoneil/ap"
)

func ExampleInverse() {
	a := ap.Assignment{1, 3, 2, 0}
	fmt.Println(ap.Inverse(a))
	// Output: [3 0 2 1]
}
