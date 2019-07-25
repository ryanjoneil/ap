package ap

import (
	"math/rand"
	"testing"
	"time"
)

const max = 100000
const M = max * 1000

func randomAP(n int) AP {
	rand.Seed(time.Now().UTC().UnixNano())
	ap := Create(n)
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			if i == j {
				ap.A[i][j] = rand.Int63n(max)
			}
		}
	}
	return ap
}

func benchSolv(b *testing.B, size int, removes int) {
	for n := 0; n < b.N; n++ {
		ap := randomAP(size)
		ap.Solve()

		// Incrementally remove assignments and re-optimize.
		for i := 0; i < removes; i++ {
			row := rand.Intn(size)
			col := ap.Col(row)
			ap.Remove(row, col, M)
			ap.Solve()
		}
	}
}

func benchCopy(b *testing.B, size int, copies int) {
	for n := 0; n < b.N; n++ {
		ap := randomAP(size)
		for i := 0; i < copies; i++ {
			ap.Copy()
		}
	}
}

func BenchmarkSolve10(b *testing.B)  { benchSolv(b, 10, 0) }
func BenchmarkSolve100(b *testing.B) { benchSolv(b, 100, 0) }
func BenchmarkSolve1K(b *testing.B)  { benchSolv(b, 1000, 0) }
func BenchmarkSolve10K(b *testing.B) { benchSolv(b, 10000, 0) }

func BenchmarkIncremental10x1(b *testing.B)   { benchSolv(b, 10, 1) }
func BenchmarkIncremental100x10(b *testing.B) { benchSolv(b, 100, 10) }
func BenchmarkIncremental1Kx100(b *testing.B) { benchSolv(b, 1000, 100) }

func BenchmarkCopy10x1K(b *testing.B)  { benchCopy(b, 10, 1000) }
func BenchmarkCopy100x1K(b *testing.B) { benchCopy(b, 100, 1000) }
func BenchmarkCopy1Kx1K(b *testing.B)  { benchCopy(b, 1000, 1000) }
