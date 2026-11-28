package klsap_test

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"runtime"
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

// liveHeap returns the bytes of live heap objects after a garbage collection.
func liveHeap() int64 {
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return int64(m.HeapAlloc)
}

// held returns held-B/subproblem for ranking c: the live heap that the first
// split adds, divided by the number of subproblems it keeps, or 0 if it keeps
// none. The first Next solves the root. The second splits it into n − 1
// children, keeps the best min(k − 1, n − 1), and returns one of them, which
// stays live until the next split. Every child of a dense problem is
// feasible, so that count is exact. The children share the root's cost rows,
// so the metric is what one more waiting subproblem costs.
func held(b *testing.B, c [][]int64, k int) float64 {
	b.Helper()

	kept := min(k-1, len(c)-1)
	if kept < 1 {
		return 0
	}
	a, err := klsap.New(c, k)
	if err != nil {
		b.Fatal(err)
	}
	if _, err := a.Next(); err != nil {
		b.Fatal(err)
	}
	before := liveHeap()
	if _, err := a.Next(); err != nil {
		b.Fatal(err)
	}
	after := liveHeap()
	runtime.KeepAlive(a)
	return float64(after-before) / float64(kept)
}

// benchDense ranks the k best assignments of a uniform dense instance of
// size n.
func benchDense(b *testing.B, n, k int) {
	b.ReportAllocs()
	c := instances.Dense(n, instances.Uniform(benchRNG(), benchK))
	h := held(b, c, k)
	for b.Loop() {
		a, err := klsap.New(c, k)
		if err != nil {
			b.Fatal(err)
		}
		rankAll(b, a, k)
	}
	// Reported after the loop, since b.Loop clears metrics reported before it.
	b.ReportMetric(h, "held-B/subproblem")
}

// BenchmarkDense leaves out n = 1000 with k = 100, which takes over a minute
// per run, so that go test -bench . stays practical. BenchmarkLargeDense has
// it.
func BenchmarkDense(b *testing.B) {
	for _, tc := range []struct{ n, k int }{{100, 10}, {100, 100}, {1000, 10}} {
		b.Run(fmt.Sprintf("n=%d/k=%d", tc.n, tc.k), func(b *testing.B) {
			benchDense(b, tc.n, tc.k)
		})
	}
}

// BenchmarkLargeDense is opt-in; see [instances.LargeEnv].
func BenchmarkLargeDense(b *testing.B) {
	instances.SkipUnlessLarge(b)
	n, k := 1000, 100
	b.Run(fmt.Sprintf("n=%d/k=%d", n, k), func(b *testing.B) {
		benchDense(b, n, k)
	})
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
