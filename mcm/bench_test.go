package mcm_test

import (
	"math/rand/v2"
	"testing"

	"github.com/ryanjoneil/ap"
	"github.com/ryanjoneil/ap/internal/instances"
	"github.com/ryanjoneil/ap/mcm"
)

// Instances are generated once with seed 1, outside the timed loop.

func BenchmarkSparse100K(b *testing.B) {
	rng := rand.New(rand.NewPCG(1, 1))
	edges := instances.Sparse(rng, 100000, 10, instances.Uniform(rng, 1))
	pairs := make([]ap.Pair, len(edges))
	for i, e := range edges {
		pairs[i] = ap.Pair{U: e.U, V: e.V}
	}
	b.ResetTimer()
	for b.Loop() {
		m, err := mcm.NewSparse(100000, 100000, pairs)
		if err != nil {
			b.Fatal(err)
		}
		m.Match()
	}
}

func BenchmarkDense2K(b *testing.B) {
	rng := rand.New(rand.NewPCG(1, 1))
	adj := make([][]bool, 2000)
	for u := range adj {
		adj[u] = make([]bool, 2000)
		for v := range adj[u] {
			adj[u][v] = rng.Float64() < 0.5
		}
	}
	b.ResetTimer()
	for b.Loop() {
		m, err := mcm.New(adj)
		if err != nil {
			b.Fatal(err)
		}
		m.Match()
	}
}
