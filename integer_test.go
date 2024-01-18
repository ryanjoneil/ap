package ap_test

import (
	"fmt"

	"github.com/ryanjoneil/ap"
)

func ExampleMaxOf() {
	i := ap.MaxOf[int64]()
	fmt.Println(i)
	// Output:
	// 9223372036854775807
}

func ExampleMinOf() {
	i := ap.MinOf[int64]()
	fmt.Println(i)
	// Output:
	// -9223372036854775808
}
