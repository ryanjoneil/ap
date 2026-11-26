package lbap_test

import (
	"math/rand/v2"
	"testing"

	"github.com/ryanjoneil/ap/internal/instances"
	"github.com/ryanjoneil/ap/lbap"
)

// Benchmark instances use the uniform class of book §4.10.1 with K = 10⁶.
// Each benchmark generates its instance once with seed 1, outside the timed
// loop.
const benchK = 1_000_000

func BenchmarkDenseUniform1K(b *testing.B) {
	rng := rand.New(rand.NewPCG(1, 1))
	c := instances.Dense(1000, instances.Uniform(rng, benchK))
	for b.Loop() {
		a, err := lbap.New(c)
		if err != nil {
			b.Fatal(err)
		}
		if _, err := a.Assign(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSparseUniform100K(b *testing.B) {
	rng := rand.New(rand.NewPCG(1, 1))
	n := 100_000
	edges := instances.Sparse(rng, n, 10, instances.Uniform(rng, benchK))
	for b.Loop() {
		a, err := lbap.NewSparse(n, edges)
		if err != nil {
			b.Fatal(err)
		}
		if _, err := a.Assign(); err != nil {
			b.Fatal(err)
		}
	}
}
