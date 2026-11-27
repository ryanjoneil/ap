package lsap_test

import (
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

func benchDense(b *testing.B, n int) {
	c := denseInstance(n)
	for b.Loop() {
		a, err := lsap.New(c)
		if err != nil {
			b.Fatal(err)
		}
		mustAssign(b, a)
	}
}

func benchSparse(b *testing.B, n, degree int) {
	edges := sparseInstance(n, degree)
	for b.Loop() {
		a, err := lsap.NewSparse(n, edges)
		if err != nil {
			b.Fatal(err)
		}
		mustAssign(b, a)
	}
}

func BenchmarkDenseUniform100(b *testing.B) { benchDense(b, 100) }
func BenchmarkDenseUniform1K(b *testing.B)  { benchDense(b, 1000) }
func BenchmarkDenseUniform5K(b *testing.B)  { benchDense(b, 5000) }

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
	benchSparse(b, 10_000, instances.BookDegree(10_000))
}

func BenchmarkSparseUniform100K(b *testing.B) { benchSparse(b, 100_000, 10) }

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
