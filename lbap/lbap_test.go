package lbap_test

import (
	"errors"
	"math/rand/v2"
	"slices"
	"strings"
	"testing"

	"github.com/ryanjoneil/ap"
	"github.com/ryanjoneil/ap/internal/bruteforce"
	"github.com/ryanjoneil/ap/lbap"
)

func randomCosts(rng *rand.Rand, n int) [][]int64 {
	c := make([][]int64, n)
	for i := range c {
		c[i] = make([]int64, n)
		for j := range c[i] {
			c[i][j] = rng.Int64N(21) - 10 // [-10, 10]
		}
	}
	return c
}

func fullMask(n int) [][]bool {
	mask := make([][]bool, n)
	for i := range mask {
		mask[i] = make([]bool, n)
		for j := range mask[i] {
			mask[i][j] = true
		}
	}
	return mask
}

func maskedEdges(c [][]int64, mask [][]bool) []ap.Edge[int64] {
	var edges []ap.Edge[int64]
	for i, row := range c {
		for j, cij := range row {
			if mask[i][j] {
				edges = append(edges, ap.Edge[int64]{U: i, V: j, Cost: cij})
			}
		}
	}
	return edges
}

// check solves a and compares it with brute force over the allowed edges in
// mask. It returns the assignment, or nil if both agree that none exists.
func check(t *testing.T, name string, a *lbap.LBAP[int64], c [][]int64, mask [][]bool) ap.Permutation {
	t.Helper()

	want, feasible := bruteforce.MinMax(c, mask)
	p, err := a.Assign()
	if !feasible {
		if !errors.Is(err, ap.ErrInfeasible) {
			t.Fatalf("%s: %v mask %v: got %v, %v; want ErrInfeasible", name, c, mask, p, err)
		}
		return nil
	}
	if err != nil {
		t.Fatalf("%s: %v mask %v: Assign: %v; want bottleneck %v", name, c, mask, err, want)
	}
	if len(p) != len(c) {
		t.Fatalf("%s: permutation %v has length %d; want %d", name, p, len(p), len(c))
	}
	seen := make([]bool, len(c))
	var got int64
	for u, v := range p {
		if v < 0 || v >= len(c) || seen[v] || !mask[u][v] {
			t.Fatalf("%s: %v mask %v: invalid permutation %v", name, c, mask, p)
		}
		seen[v] = true
		if u == 0 || c[u][v] > got {
			got = c[u][v]
		}
	}
	if got != want || a.Cost() != want {
		t.Fatalf("%s: %v mask %v: %v has bottleneck %v, Cost() = %v; want %v",
			name, c, mask, p, got, a.Cost(), want)
	}
	return p
}

// Review Focus 2.
func TestBruteForce(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 1))
	for range 2000 {
		n := 1 + rng.IntN(6)
		c := randomCosts(rng, n)
		d, err := lbap.New(c)
		if err != nil {
			t.Fatal(err)
		}
		check(t, "dense", d, c, fullMask(n))

		mask := fullMask(n)
		for i := range mask {
			for j := range mask[i] {
				mask[i][j] = rng.Float64() < 0.6
			}
		}
		s, err := lbap.NewSparse(n, maskedEdges(c, mask))
		if err != nil {
			t.Fatal(err)
		}
		check(t, "sparse", s, c, mask)
	}
}

// Review Focus 3.
func TestIsolated(t *testing.T) {
	noRow, _ := lbap.NewSparse(2, []ap.Edge[int64]{{U: 0, V: 0, Cost: 1}, {U: 0, V: 1, Cost: 1}})
	noCol, _ := lbap.NewSparse(2, []ap.Edge[int64]{{U: 0, V: 0, Cost: 1}, {U: 1, V: 0, Cost: 1}})
	noEdges, _ := lbap.NewSparse[int64](3, nil)
	for _, a := range []*lbap.LBAP[int64]{noRow, noCol, noEdges} {
		if _, err := a.Assign(); !errors.Is(err, ap.ErrInfeasible) {
			t.Errorf("got %v", err)
		}
	}
}

// Review Focus 4.
func TestRemoveWholeRow(t *testing.T) {
	a, _ := lbap.NewSparse(2, []ap.Edge[int64]{
		{U: 0, V: 0, Cost: 1}, {U: 0, V: 1, Cost: 2}, {U: 1, V: 0, Cost: 3}, {U: 1, V: 1, Cost: 4},
	})
	_, _ = a.Assign()
	b := a.Copy()
	a.Remove(0, 0)
	a.Remove(0, 1)
	if _, err := a.Assign(); !errors.Is(err, ap.ErrInfeasible) {
		t.Errorf("got %v", err)
	}
	// [1 0] has max(2, 3) = 3, which beats [0 1] at max(1, 4) = 4.
	if p, err := b.Assign(); err != nil || b.Cost() != 3 || !slices.Equal(p, ap.Permutation{1, 0}) {
		t.Errorf("copy: %v %v %v", p, err, b.Cost())
	}
}

// TestRemoveIncremental repeatedly removes an assigned edge, plus an arbitrary
// other edge (assigned or not, possibly already removed) and some no-ops, and
// compares each re-solve with brute force on the shrinking mask.
func TestRemoveIncremental(t *testing.T) {
	rng := rand.New(rand.NewPCG(2, 2))
	for trial := range 300 {
		n := 2 + rng.IntN(5)
		c := randomCosts(rng, n)
		if trial%2 == 1 {
			for i := range c { // few distinct costs, so many ties
				for j := range c[i] {
					c[i][j] = rng.Int64N(3)
				}
			}
		}
		mask := fullMask(n)
		a, err := lbap.New(c)
		if err != nil {
			t.Fatal(err)
		}
		for {
			p := check(t, "incremental", a, c, mask)
			if p == nil {
				break
			}
			u := rng.IntN(n)
			a.Remove(u, p[u])
			mask[u][p[u]] = false
			a.Remove(u, p[u]) // already removed: a no-op
			a.Remove(-1, 0)
			a.Remove(0, n)
			a.Remove(n, 0)
			if p := check(t, "after assigned", a, c, mask); p == nil {
				break
			}
			ru, rv := rng.IntN(n), rng.IntN(n)
			a.Remove(ru, rv)
			mask[ru][rv] = false
		}
	}
}

// Removing an edge outside the assignment keeps the cached solution.
func TestRemoveUnassignedKeepsSolution(t *testing.T) {
	c := [][]int64{{1, 9}, {9, 1}}
	a, _ := lbap.New(c)
	before, _ := a.Assign()
	a.Remove(0, 1)
	a.Remove(1, 0)
	after, err := a.Assign()
	if err != nil || a.Cost() != 1 || !slices.Equal(before, after) {
		t.Fatalf("got %v %v %v; want %v", after, err, a.Cost(), before)
	}
}

// Review Focus 1.
func TestCopyIndependent(t *testing.T) {
	c := [][]int64{{4, 1, 3}, {2, 0, 5}, {3, 2, 2}}
	maskA := fullMask(3)
	a, _ := lbap.New(c)
	check(t, "a", a, c, maskA)
	a.Remove(0, 0) // The copy must inherit this removal.
	maskA[0][0] = false
	b := a.Copy()
	maskB := fullMask(3)
	maskB[0][0] = false

	pa, _ := a.Assign()
	costA := a.Cost()
	pb := check(t, "b", b, c, maskB)
	for u, v := range pb {
		b.Remove(u, v)
		maskB[u][v] = false
	}
	if got, _ := a.Assign(); !slices.Equal(got, pa) || a.Cost() != costA {
		t.Fatalf("original changed by copy: %v %v; want %v %v", got, a.Cost(), pa, costA)
	}
	check(t, "a after b", a, c, maskA)
	check(t, "b", b, c, maskB)

	pb, _ = b.Assign()
	costB := b.Cost()
	for u, v := range pa {
		a.Remove(u, v)
		maskA[u][v] = false
	}
	check(t, "a", a, c, maskA)
	if got, _ := b.Assign(); !slices.Equal(got, pb) || b.Cost() != costB {
		t.Fatalf("copy changed by original: %v %v; want %v %v", got, b.Cost(), pb, costB)
	}
	check(t, "b after a", b, c, maskB)
}

func TestDeterministic(t *testing.T) {
	c := make([][]int64, 6)
	for i := range c {
		c[i] = []int64{1, 1, 1, 1, 1, 1}
	}
	assign := func() ap.Permutation {
		a, err := lbap.New(c)
		if err != nil {
			t.Fatal(err)
		}
		p, err := a.Assign()
		if err != nil {
			t.Fatal(err)
		}
		return p
	}
	first := assign()
	for run := range 100 {
		if p := assign(); !slices.Equal(first, p) {
			t.Fatalf("run %v: got %v; first run got %v", run, p, first)
		}
	}
}

func TestNewErrors(t *testing.T) {
	cases := []struct {
		name string
		f    func() error
		want string
	}{
		{"empty", func() error { _, e := lbap.New([][]int64{}); return e }, "empty"},
		{"empty row", func() error { _, e := lbap.New([][]int64{{}}); return e }, "empty"},
		{"ragged", func() error { _, e := lbap.New([][]int64{{1, 2}, {3}}); return e }, "not square"},
		{"not square", func() error { _, e := lbap.New([][]int64{{1, 2}, {3, 4}, {5, 6}}); return e }, "not square"},
		{"sparse empty", func() error { _, e := lbap.NewSparse[int64](0, nil); return e }, "empty"},
		{"out of range", func() error {
			_, e := lbap.NewSparse(2, []ap.Edge[int64]{{U: 0, V: 2, Cost: 1}})
			return e
		}, "out of range"},
		{"negative", func() error {
			_, e := lbap.NewSparse(2, []ap.Edge[int64]{{U: -1, V: 0, Cost: 1}})
			return e
		}, "out of range"},
		{"duplicate", func() error {
			_, e := lbap.NewSparse(2, []ap.Edge[int64]{{U: 0, V: 1, Cost: 1}, {U: 0, V: 1, Cost: 2}})
			return e
		}, "duplicate"},
	}
	for _, c := range cases {
		err := c.f()
		if err == nil || !strings.HasPrefix(err.Error(), "lbap: ") ||
			!strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: got %v", c.name, err)
		}
	}
}

func TestNewCopiesInput(t *testing.T) {
	c := [][]int64{{3, 1}, {2, 4}}
	a, _ := lbap.New(c)
	a.Remove(0, 1)
	_, _ = a.Assign()
	if !slices.Equal(c[0], []int64{3, 1}) || !slices.Equal(c[1], []int64{2, 4}) {
		t.Fatalf("costs modified: %v", c)
	}
	c[0][0] = 100 // later changes don't affect the solver
	if _, err := a.Assign(); err != nil || a.Cost() != 4 {
		t.Fatalf("got %v %v; want cost 4", err, a.Cost())
	}

	edges := []ap.Edge[int64]{{U: 1, V: 1, Cost: 4}, {U: 0, V: 1, Cost: 2}, {U: 0, V: 0, Cost: 1}, {U: 1, V: 0, Cost: 3}}
	orig := slices.Clone(edges)
	s, _ := lbap.NewSparse(2, edges)
	s.Remove(0, 0)
	_, _ = s.Assign()
	if !slices.Equal(edges, orig) {
		t.Fatalf("edges modified: %v", edges)
	}
}

// Costs are only compared, so the extremes of T are fine.
func TestExtremeCosts(t *testing.T) {
	a, err := lbap.New([][]int8{{-128, 127}, {127, -128}})
	if err != nil {
		t.Fatal(err)
	}
	if p, err := a.Assign(); err != nil || a.Cost() != -128 || !slices.Equal(p, ap.Permutation{0, 1}) {
		t.Fatalf("got %v %v %v", p, err, a.Cost())
	}
}
