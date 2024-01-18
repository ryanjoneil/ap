package lsap_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/ryanjoneil/ap/lsap"
)

func randomMatrix(n int) [][]int {
	rng := rand.New(rand.NewSource(time.Now().UTC().UnixNano()))
	A := make([][]int, n)
	for i := 0; i < n; i++ {
		Ai := make([]int, n)
		for j := 0; j < n; j++ {
			if i == j {
				Ai[j] = rng.Intn(100000)
			}
		}
		A[i] = Ai
	}
	return A
}

func benchSolve(b *testing.B, size int, removes int) {
	b.Helper()

	matrices := [][][]int{}
	for n := 0; n < b.N; n++ {
		matrices = append(matrices, randomMatrix(size))
	}

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		ap := lsap.New(matrices[n])
		assign := ap.Assign()

		// Incrementally remove assignments and re-optimize.
		for i := 0; i < removes; i++ {
			u := rand.Intn(size)
			v := assign[u]
			ap.Remove(u, v)
			assign = ap.Assign()
		}
	}
}

func benchCopy(b *testing.B, size int, copies int) {
	b.Helper()

	matrices := [][][]int{}
	for n := 0; n < b.N; n++ {
		matrices = append(matrices, randomMatrix(size))
	}

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		ap := lsap.New(matrices[n])
		for i := 0; i < copies; i++ {
			ap.Copy()
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
