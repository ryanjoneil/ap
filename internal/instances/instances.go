// Package instances generates the benchmark instance classes for the linear
// sum assignment problem described in
//
//	R. Burkard, M. Dell'Amico and S. Martello, Assignment Problems (SIAM,
//	2012), Section 4.10.1 "Classes of instances".
//
// The classes are:
//
//   - Uniform: integers uniformly random in [0, K].
//   - Geometric: two sets X, Y of n points with integer coordinates in
//     [1, K] x [1, K]; c_ij is the truncated Euclidean distance between the
//     i-th point of X and the j-th point of Y. Used by Goldberg and Kennedy
//     [340].
//   - Two-cost: cost 1 with probability 0.5 or 10^6 with probability 0.5,
//     derived from an analogous class of Goldberg and Kennedy [340].
//   - Machol-Wien: c_ij = (i-1)(j-1) for i, j = 1..n ([488, 489]).
//   - Sparse: randomly selected entries in each row; the book uses 2 log n
//     entries per row.
//
// References (numbers are those of the book's bibliography):
//
//	[340] A.V. Goldberg and R. Kennedy. An efficient cost scaling algorithm for
//	      the assignment problem. Math. Program., 71:153-177, 1995.
//	[488] R.E. Machol and M. Wien. A hard assignment problem. Oper. Res.,
//	      24:190-192, 1976.
//	[489] R.E. Machol and M. Wien. Errata. Oper. Res., 25:364, 1977.
//
// All generators take a *rand.Rand from math/rand/v2, so instances are
// reproducible from a seed.
package instances

import (
	"math"
	"math/bits"
	"math/rand/v2"

	"github.com/ryanjoneil/ap"
)

// Class gives the cost of assigning row i to column j (0-based).
//
// Uniform and TwoCost classes draw from their generator on every call, so a
// Class must be evaluated in a fixed order to be reproducible; Dense and
// Sparse do so.
type Class func(i, j int) int64

// Uniform returns the uniformly random class of book Section 4.10.1: each
// cost is an integer drawn uniformly from [0, k].
func Uniform(rng *rand.Rand, k int64) Class {
	return func(_, _ int) int64 { return rng.Int64N(k + 1) }
}

// Geometric returns the geometric class of book Section 4.10.1, as used by
// Goldberg and Kennedy [340]. It draws n points X and n points Y (all X, then
// all Y) with integer coordinates in [1, k] x [1, k] immediately; the
// returned Class is then a pure function. Cost (i, j) is the truncated
// Euclidean distance between X[i] and Y[j].
func Geometric(rng *rand.Rand, n int, k int64) Class {
	type pt struct{ x, y int64 }
	draw := func() []pt {
		ps := make([]pt, n)
		for i := range ps {
			ps[i] = pt{rng.Int64N(k) + 1, rng.Int64N(k) + 1}
		}
		return ps
	}
	xs := draw()
	ys := draw()
	return func(i, j int) int64 {
		dx := float64(xs[i].x - ys[j].x)
		dy := float64(xs[i].y - ys[j].y)
		return int64(math.Sqrt(dx*dx + dy*dy)) // truncation toward zero
	}
}

// TwoCost returns the two-cost class of book Section 4.10.1 (derived from
// Goldberg and Kennedy [340]): cost 1 or 1,000,000, each with probability
// 0.5.
func TwoCost(rng *rand.Rand) Class {
	return func(_, _ int) int64 {
		if rng.IntN(2) == 0 {
			return 1
		}
		return 1_000_000
	}
}

// MacholWien returns the deterministic hard class of Machol and Wien
// ([488, 489]; book Section 4.10.1): c_ij = (i-1)(j-1) for 1-based i, j,
// which is i*j for the 0-based indices used here.
func MacholWien() Class {
	return func(i, j int) int64 { return int64(i) * int64(j) }
}

// Dense returns the n x n matrix of costs c(i, j). It calls c in row-major
// order (i outer, j inner), which makes stateful classes deterministic.
func Dense(n int, c Class) [][]int64 {
	m := make([][]int64, n)
	for i := range m {
		m[i] = make([]int64, n)
		for j := range m[i] {
			m[i][j] = c(i, j)
		}
	}
	return m
}

// Sparse returns an n x n sparse instance with exactly min(degree, n)
// distinct columns in each row, with costs from c. Edges are emitted row by
// row, with columns in the order drawn (not sorted); c is called once per
// edge in that order.
//
// The book's sparse class (Section 4.10.1) selects entries at random in each
// row. As a departure from the book, Sparse always includes the entries of a
// hidden random permutation, so that a perfect matching is guaranteed to
// exist.
//
// Sparse panics if n < 0 or degree <= 0; it is a helper for tests and
// benchmarks, not library API.
func Sparse(rng *rand.Rand, n, degree int, c Class) []ap.Edge[int64] {
	if n < 0 || degree <= 0 {
		panic("instances: Sparse requires n >= 0 and degree > 0")
	}
	if degree > n {
		degree = n
	}
	perm := rng.Perm(n)
	edges := make([]ap.Edge[int64], 0, n*degree)
	add := func(i, j int) {
		edges = append(edges, ap.Edge[int64]{U: i, V: j, Cost: c(i, j)})
	}
	if degree > n/2 {
		// Dense rows: partial Fisher-Yates over all columns, with the
		// permutation column swapped to the front.
		cols := make([]int, n)
		pos := make([]int, n) // pos[col] = index of col in cols
		for j := range cols {
			cols[j], pos[j] = j, j
		}
		swap := func(a, b int) {
			cols[a], cols[b] = cols[b], cols[a]
			pos[cols[a]], pos[cols[b]] = a, b
		}
		for i := 0; i < n; i++ {
			swap(0, pos[perm[i]])
			for k := 1; k < degree; k++ {
				swap(k, k+rng.IntN(n-k))
			}
			for k := 0; k < degree; k++ {
				add(i, cols[k])
			}
		}
		return edges
	}
	// Sparse rows: rejection sampling, with a per-row stamp for O(1) membership.
	stamp := make([]int, n)
	for i := range stamp {
		stamp[i] = -1
	}
	for i := 0; i < n; i++ {
		stamp[perm[i]] = i
		add(i, perm[i])
		for k := 1; k < degree; k++ {
			j := rng.IntN(n)
			for stamp[j] == i {
				j = rng.IntN(n)
			}
			stamp[j] = i
			add(i, j)
		}
	}
	return edges
}

// BookDegree returns 2*ceil(log2 n), the per-row density of the book's sparse
// class (Section 4.10.1: "2 log n entries"). The book does not state the
// logarithm base or rounding; this reading is ours. It is at least 1.
func BookDegree(n int) int {
	if n <= 1 {
		return 1
	}
	return 2 * bits.Len(uint(n-1))
}
