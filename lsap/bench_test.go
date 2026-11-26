package lsap_test

import (
	"math/rand"
	"testing"

	"github.com/ryanjoneil/ap/lsap"
)

func randomMatrix(rng *rand.Rand, n int) [][]int {
	A := make([][]int, n)
	for i := 0; i < n; i++ {
		Ai := make([]int, n)
		for j := 0; j < n; j++ {
			Ai[j] = rng.Intn(100000)
		}
		A[i] = Ai
	}
	return A
}

func benchSolve(b *testing.B, size int, removes int) {
	b.Helper()

	rng := rand.New(rand.NewSource(1))
	matrices := [][][]int{}
	for n := 0; n < b.N; n++ {
		matrices = append(matrices, randomMatrix(rng, size))
	}

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		a, err := lsap.New(matrices[n])
		if err != nil {
			b.Fatal(err)
		}
		assign, err := a.Assign()
		if err != nil {
			b.Fatal(err)
		}

		// Incrementally remove assignments and re-optimize.
		for i := 0; i < removes; i++ {
			u := rng.Intn(size)
			v := assign[u]
			a.Remove(u, v)
			if assign, err = a.Assign(); err != nil {
				b.Fatal(err)
			}
		}
	}
}

func benchCopy(b *testing.B, size int, copies int) {
	b.Helper()

	rng := rand.New(rand.NewSource(1))
	matrices := [][][]int{}
	for n := 0; n < b.N; n++ {
		matrices = append(matrices, randomMatrix(rng, size))
	}

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		a, err := lsap.New(matrices[n])
		if err != nil {
			b.Fatal(err)
		}
		for i := 0; i < copies; i++ {
			a.Copy()
		}
	}
}

func BenchmarkSolve100(b *testing.B) { benchSolve(b, 100, 0) }
func BenchmarkSolve1K(b *testing.B)  { benchSolve(b, 1000, 0) }
func BenchmarkSolve10K(b *testing.B) { benchSolve(b, 10000, 0) }

func BenchmarkIncremental100x10(b *testing.B) { benchSolve(b, 100, 10) }
func BenchmarkIncremental1Kx100(b *testing.B) { benchSolve(b, 1000, 100) }

func BenchmarkCopy100x1K(b *testing.B) { benchCopy(b, 100, 1000) }
func BenchmarkCopy1Kx1K(b *testing.B)  { benchCopy(b, 1000, 1000) }
