package main

import (
	"fmt"

	"github.com/ryanjoneil/ap"
)

func main() {
	ap := ap.New([][]int64{
		{25, 40, 9, 15},
		{12, 33, 22, 31},
		{18, 37, 18, 30},
		{32, 11, 12, 26},
	})
	ap.Solve()

	fmt.Printf("Z =\t%v\n", ap.Z)
	fmt.Printf("U =\t%v\n", ap.U)
	fmt.Printf("V =\t%v\n", ap.V)
	fmt.Println()

	fmt.Print("RC =\t")
	for i := 0; i < ap.Size; i++ {
		if i > 0 {
			fmt.Print("\t")
		}
		for j := 0; j < ap.Size; j++ {
			fmt.Printf("% 3d", ap.RC(i, j))
		}
		fmt.Println()
	}
	fmt.Println()

	fmt.Print("X =\t")
	for i := 0; i < ap.Size; i++ {
		if i > 0 {
			fmt.Print("\t")
		}
		for j := 0; j < ap.Size; j++ {
			if ap.X(i, j) {
				fmt.Print("  1")
			} else {
				fmt.Print("  -")
			}
		}
		fmt.Println()
	}
}
