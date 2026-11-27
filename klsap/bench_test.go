package klsap_test

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/ryanjoneil/ap"
	"github.com/ryanjoneil/ap/internal/instances"
	"github.com/ryanjoneil/ap/klsap"
)

// Benchmark instances use the uniform class of book §4.10.1 with K = 10⁶.
// Each benchmark generates its instance once with seed 1, outside the timed
// loop.
const benchK = 1_000_000

func benchRNG() *rand.Rand { return rand.New(rand.NewPCG(1, 1)) }

// rankAll calls Next until it returns an error, which must be
// ap.ErrInfeasible after k assignments.
func rankAll(b *testing.B, a *klsap.KLSAP[int64], k int) {
	b.Helper()

	for r := 0; ; r++ {
		if _, err := a.Next(); err != nil {
			if !errors.Is(err, ap.ErrInfeasible) || r != k {
				b.Fatalf("Next %d: %v", r, err)
			}
			return
		}
	}
}

func BenchmarkDense(b *testing.B) {
	for _, n := range []int{100, 1000} {
		c := instances.Dense(n, instances.Uniform(benchRNG(), benchK))
		for _, k := range []int{10, 100} {
			b.Run(fmt.Sprintf("n=%d/k=%d", n, k), func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					a, err := klsap.New(c, k)
					if err != nil {
						b.Fatal(err)
					}
					rankAll(b, a, k)
				}
			})
		}
	}
}

func BenchmarkSparse(b *testing.B) {
	n, k := 1000, 10
	rng := benchRNG()
	edges := instances.Sparse(rng, n, instances.BookDegree(n), instances.Uniform(rng, benchK))
	b.Run(fmt.Sprintf("n=%d/k=%d", n, k), func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			a, err := klsap.NewSparse(n, edges, k)
			if err != nil {
				b.Fatal(err)
			}
			rankAll(b, a, k)
		}
	})
}
