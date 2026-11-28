package lsap_test

import (
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/ryanjoneil/ap"
	"github.com/ryanjoneil/ap/internal/instances"
	"github.com/ryanjoneil/ap/lsap"
)

// Benchmark instances use the uniform class of book §4.10.1 with K = 10⁶.
// Each benchmark generates its instance once with seed 1, outside the timed
// loop. b.Loop runs the benchmark function body once, so the setup is never
// repeated.
const benchK = 1_000_000

func benchRNG() *rand.Rand { return rand.New(rand.NewPCG(1, 1)) }

func denseInstance(n int) [][]int64 {
	rng := benchRNG()
	return instances.Dense(n, instances.Uniform(rng, benchK))
}

func sparseInstance(n, degree int) []ap.Edge[int64] {
	rng := benchRNG()
	return instances.Sparse(rng, n, degree, instances.Uniform(rng, benchK))
}

// removeRows returns count rows of an n-row problem, chosen once so that every
// iteration removes the same assigned edges.
func removeRows(n, count int) []int {
	rng := rand.New(rand.NewPCG(2, 2))
	rows := make([]int, count)
	for i := range rows {
		rows[i] = rng.IntN(n)
	}
	return rows
}

func mustAssign(b *testing.B, a *lsap.LSAP[int64]) ap.Permutation {
	b.Helper()

	p, err := a.Assign()
	if err != nil {
		b.Fatal(err)
	}
	return p
}

// benchIncremental solves a, then removes the assigned edge of each row in
// rows in turn, re-solving after each removal.
func benchIncremental(b *testing.B, a *lsap.LSAP[int64], rows []int) {
	b.Helper()

	p := mustAssign(b, a)
	for _, u := range rows {
		a.Fix(u, p[u], false)
		p = mustAssign(b, a)
	}
}

func benchDense(b *testing.B, c [][]int64) {
	for b.Loop() {
		a, err := lsap.New(c)
		if err != nil {
			b.Fatal(err)
		}
		mustAssign(b, a)
	}
}

func benchSparse(b *testing.B, n int, edges []ap.Edge[int64]) {
	for b.Loop() {
		a, err := lsap.NewSparse(n, edges)
		if err != nil {
			b.Fatal(err)
		}
		mustAssign(b, a)
	}
}

func BenchmarkDenseUniform100(b *testing.B) { benchDense(b, denseInstance(100)) }
func BenchmarkDenseUniform1K(b *testing.B)  { benchDense(b, denseInstance(1000)) }

func BenchmarkDenseIncremental1Kx100(b *testing.B) {
	c := denseInstance(1000)
	rows := removeRows(1000, 100)
	for b.Loop() {
		a, err := lsap.New(c)
		if err != nil {
			b.Fatal(err)
		}
		benchIncremental(b, a, rows)
	}
}

func BenchmarkDenseCopy1Kx1K(b *testing.B) {
	c := denseInstance(1000)
	for b.Loop() {
		a, err := lsap.New(c)
		if err != nil {
			b.Fatal(err)
		}
		mustAssign(b, a)
		for range 1000 {
			a.Copy()
		}
	}
}

func BenchmarkSparseUniform10K(b *testing.B) {
	n := 10_000
	benchSparse(b, n, sparseInstance(n, instances.BookDegree(n)))
}

func BenchmarkSparseIncremental10Kx100(b *testing.B) {
	n := 10_000
	edges := sparseInstance(n, instances.BookDegree(n))
	rows := removeRows(n, 100)
	for b.Loop() {
		a, err := lsap.NewSparse(n, edges)
		if err != nil {
			b.Fatal(err)
		}
		benchIncremental(b, a, rows)
	}
}

// largeClass is a book §4.10.1 class timed by the large benchmarks, at the
// sizes v0.5's external comparison used. Each instance uses seed 1 and
// K = 10⁶, as that comparison's instances did, so times stay comparable.
type largeClass struct {
	name  string
	sizes []int
	class func(rng *rand.Rand, n int) instances.Class
}

var (
	uniformClass = largeClass{"uniform", []int{1000, 5000},
		func(rng *rand.Rand, _ int) instances.Class { return instances.Uniform(rng, benchK) }}
	geometricClass = largeClass{"geometric", []int{1000, 5000},
		func(rng *rand.Rand, n int) instances.Class { return instances.Geometric(rng, n, benchK) }}
	twoCostClass = largeClass{"twocost", []int{1000, 5000},
		func(rng *rand.Rand, _ int) instances.Class { return instances.TwoCost(rng) }}
	// Machol–Wien stops at n = 2,000: it is by far the slowest class, at
	// about 24 s per solve there.
	macholWienClass = largeClass{"macholwien", []int{1000, 2000},
		func(*rand.Rand, int) instances.Class { return instances.MacholWien() }}
)

// BenchmarkLargeDense is opt-in; see [instances.LargeEnv].
func BenchmarkLargeDense(b *testing.B) {
	instances.SkipUnlessLarge(b)
	for _, lc := range []largeClass{uniformClass, geometricClass, twoCostClass, macholWienClass} {
		for _, n := range lc.sizes {
			b.Run(fmt.Sprintf("class=%s/n=%d", lc.name, n), func(b *testing.B) {
				benchDense(b, instances.Dense(n, lc.class(benchRNG(), n)))
			})
		}
	}
}

// BenchmarkLargeSparse is opt-in; see [instances.LargeEnv]. Its instances
// have n = 100,000 and 10 edges per row, about 10⁶ edges.
func BenchmarkLargeSparse(b *testing.B) {
	instances.SkipUnlessLarge(b)
	const n, degree = 100_000, 10
	for _, lc := range []largeClass{uniformClass, geometricClass} {
		b.Run(fmt.Sprintf("class=%s/n=%d", lc.name, n), func(b *testing.B) {
			rng := benchRNG()
			c := lc.class(rng, n)
			benchSparse(b, n, instances.Sparse(rng, n, degree, c))
		})
	}
}
