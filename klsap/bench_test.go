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
// Each sub-benchmark generates its instance with seed 1 when it runs, outside
// the timed loop. b.Loop runs the sub-benchmark function body once, so the
// setup is never repeated.
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

// BenchmarkDense leaves out n = 1000 with k = 100, which takes over a minute
// per run, so that go test -bench . stays practical.
func BenchmarkDense(b *testing.B) {
	for _, tc := range []struct{ n, k int }{{100, 10}, {100, 100}, {1000, 10}} {
		b.Run(fmt.Sprintf("n=%d/k=%d", tc.n, tc.k), func(b *testing.B) {
			b.ReportAllocs()
			c := instances.Dense(tc.n, instances.Uniform(benchRNG(), benchK))
			for b.Loop() {
				a, err := klsap.New(c, tc.k)
				if err != nil {
					b.Fatal(err)
				}
				rankAll(b, a, tc.k)
			}
		})
	}
}

func BenchmarkSparse(b *testing.B) {
	n, k := 1000, 10
	b.Run(fmt.Sprintf("n=%d/k=%d", n, k), func(b *testing.B) {
		b.ReportAllocs()
		rng := benchRNG()
		edges := instances.Sparse(rng, n, instances.BookDegree(n), instances.Uniform(rng, benchK))
		for b.Loop() {
			a, err := klsap.NewSparse(n, edges, k)
			if err != nil {
				b.Fatal(err)
			}
			rankAll(b, a, k)
		}
	})
}
