package lsap_test

import (
	"errors"
	"math"
	"math/rand"
	"reflect"
	"testing"

	"github.com/ryanjoneil/ap"
	"github.com/ryanjoneil/ap/internal/bruteforce"
	"github.com/ryanjoneil/ap/lsap"
)

// maskedEdges returns the entries of c with mask[i][j] == true as edges.
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

func newSparse(t *testing.T, n int, edges []ap.Edge[int64]) *lsap.LSAP[int64] {
	t.Helper()

	a, err := lsap.NewSparse(n, edges)
	if err != nil {
		t.Fatalf("NewSparse(%v, %v): %v", n, edges, err)
	}
	return a
}

func TestSparseMatchesBruteForce(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	for trial := 0; trial < 2000; trial++ {
		n := 1 + rng.Intn(6)
		c := randomCosts(rng, n)
		mask := fullMask(n)
		for i := range mask {
			for j := range mask[i] {
				mask[i][j] = rng.Float64() < 0.6
			}
		}
		a := newSparse(t, n, maskedEdges(c, mask))
		if p := checkAssignMask(t, "sparse", a, c, mask); p != nil {
			checkDuals(t, "sparse", a, c, mask, p)
		}
	}
}

// Review Focus 3.
func TestSparseIsolated(t *testing.T) {
	noRow, _ := lsap.NewSparse(2, []ap.Edge[int64]{{U: 0, V: 0, Cost: 1}, {U: 0, V: 1, Cost: 1}})
	noCol, _ := lsap.NewSparse(2, []ap.Edge[int64]{{U: 0, V: 0, Cost: 1}, {U: 1, V: 0, Cost: 1}})
	for _, a := range []*lsap.LSAP[int64]{noRow, noCol} {
		if _, err := a.Assign(); !errors.Is(err, ap.ErrInfeasible) {
			t.Errorf("got %v", err)
		}
	}
}

// Review Focus 4.
func TestSparseRemoveWholeRow(t *testing.T) {
	a, _ := lsap.NewSparse(2, []ap.Edge[int64]{
		{U: 0, V: 0, Cost: 1}, {U: 0, V: 1, Cost: 2}, {U: 1, V: 0, Cost: 3}, {U: 1, V: 1, Cost: 4},
	})
	_, _ = a.Assign()
	b := a.Copy()
	a.Remove(0, 0)
	a.Remove(0, 1)
	if _, err := a.Assign(); !errors.Is(err, ap.ErrInfeasible) {
		t.Errorf("got %v", err)
	}
	if _, err := b.Assign(); err != nil || b.Cost() != 5 {
		t.Errorf("copy: %v %v", err, b.Cost())
	}
}

// Review Focus 1.
func TestSparseCopyIndependent(t *testing.T) {
	c := [][]int64{{4, 1, 3}, {2, 0, 5}, {3, 2, 2}}
	a := newSparse(t, 3, fullEdges(c))
	_, _ = a.Assign()
	a.Remove(0, 0) // Not assigned; the copy must inherit this removal.
	b := a.Copy()
	b.Remove(0, 1)
	b.Remove(1, 1)
	pa, _ := a.Assign()
	if a.Cost() != 5 || !reflect.DeepEqual(pa, ap.Permutation{1, 0, 2}) {
		t.Errorf("original changed: %v %v", pa, a.Cost())
	}
	if _, ok := a.ReducedCost(1, 1); !ok {
		t.Error("Remove on the copy removed the edge from the original")
	}
	a.Remove(2, 2)
	if _, ok := b.ReducedCost(2, 2); !ok {
		t.Error("Remove on the original removed the edge from the copy")
	}
	if _, ok := b.ReducedCost(0, 0); ok {
		t.Error("copy lost a removal made before Copy")
	}
	pb, _ := b.Assign()
	want, _ := bruteforce.MinSum(c, [][]bool{{false, false, true}, {true, false, true}, {true, true, true}})
	if b.Cost() != want {
		t.Errorf("copy: %v cost %v, want %v", pb, b.Cost(), want)
	}
	// The original still has (0, 1) and (1, 1), but not (0, 0) or (2, 2).
	pa, _ = a.Assign()
	want, _ = bruteforce.MinSum(c, [][]bool{{false, true, true}, {true, true, true}, {true, true, false}})
	if a.Cost() != want {
		t.Errorf("original: %v cost %v, want %v", pa, a.Cost(), want)
	}
}

// Remove each assigned edge in turn, re-solve, and compare with brute force on
// the shrinking mask.
func TestSparseRemoveIncremental(t *testing.T) {
	for trial := 0; trial < 300; trial++ {
		rng := rand.New(rand.NewSource(int64(trial)))
		n := 2 + rng.Intn(5)
		c := randomCosts(rng, n)
		runIncremental(t, "sparse", newSparse(t, n, fullEdges(c)), c, rng, false)
	}
}

func TestSparseDeterministic(t *testing.T) {
	c := [][]int64{
		{1, 2, 1, 0},
		{1, 2, 2, 0},
		{0, 0, 0, 2},
		{1, 1, 1, 0},
	}

	assign := func() ap.Permutation {
		p, err := newSparse(t, 4, fullEdges(c)).Assign()
		if err != nil {
			t.Fatal(err)
		}
		return p
	}

	first := assign()
	for run := 0; run < 100; run++ {
		if p := assign(); !reflect.DeepEqual(first, p) {
			t.Fatalf("run %v: got %v; first run got %v", run, p, first)
		}
	}
}

func TestNewSparseErrors(t *testing.T) {
	for name, tc := range map[string]struct {
		n     int
		edges []ap.Edge[int64]
	}{
		"n = 0":        {0, nil},
		"out of range": {2, []ap.Edge[int64]{{U: 0, V: 2, Cost: 1}}},
		"negative":     {2, []ap.Edge[int64]{{U: -1, V: 0, Cost: 1}}},
		"duplicate":    {2, []ap.Edge[int64]{{U: 0, V: 1, Cost: 1}, {U: 0, V: 1, Cost: 2}}},
		"too large":    {math.MaxInt32 + 1, nil},
	} {
		if _, err := lsap.NewSparse(tc.n, tc.edges); err == nil {
			t.Errorf("%s: want error", name)
		}
	}
	// 4·(2+1)·11 = 132 > 127.
	if _, err := lsap.NewSparse(2, []ap.Edge[int8]{{U: 0, V: 0, Cost: 11}}); err == nil {
		t.Error("int8 past bound: want error")
	}
	if _, err := lsap.NewSparse(2, []ap.Edge[int8]{{U: 0, V: 0, Cost: -11}}); err == nil {
		t.Error("int8 past bound (negative): want error")
	}
	// 4·(2+1)·10 = 120 ≤ 127.
	if _, err := lsap.NewSparse(2, []ap.Edge[int8]{{U: 0, V: 0, Cost: 10}, {U: 1, V: 1, Cost: -10}}); err != nil {
		t.Errorf("int8 within bound: %v", err)
	}
}

func TestNewSparseCopiesInput(t *testing.T) {
	edges := []ap.Edge[int64]{{U: 1, V: 1, Cost: 4}, {U: 0, V: 0, Cost: 1}, {U: 0, V: 1, Cost: 2}}
	orig := append([]ap.Edge[int64](nil), edges...)
	a := newSparse(t, 2, edges)
	a.Remove(0, 0)
	_, _ = a.Assign()
	if !reflect.DeepEqual(edges, orig) {
		t.Errorf("input modified: %v", edges)
	}
}

func TestSparseReducedCostMissing(t *testing.T) {
	a := newSparse(t, 2, []ap.Edge[int64]{{U: 0, V: 0, Cost: 1}, {U: 1, V: 1, Cost: 1}})
	if _, err := a.Assign(); err != nil {
		t.Fatal(err)
	}
	for _, e := range [][2]int{{0, 1}, {2, 0}, {-1, 0}, {0, 2}} {
		if _, ok := a.ReducedCost(e[0], e[1]); ok {
			t.Errorf("ReducedCost(%v, %v): want ok == false", e[0], e[1])
		}
	}
	a.Remove(0, 1) // missing: a no-op
	a.Remove(5, 5) // out of range: a no-op
	if _, err := a.Assign(); err != nil || a.Cost() != 2 {
		t.Errorf("got %v, cost %v", err, a.Cost())
	}
}

// Random int8 instances with costs at the overflow bound match brute force
// after random removals, so the bound leaves enough room for the duals and
// path lengths.
func TestSparseOverflowBoundInt8(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	for trial := 0; trial < 2000; trial++ {
		n := 1 + rng.Intn(4)
		bound := math.MaxInt8 / (4 * (n + 1))
		c := make([][]int64, n)
		var edges []ap.Edge[int8]
		mask := fullMask(n)
		for i := range c {
			c[i] = make([]int64, n)
			for j := range c[i] {
				c[i][j] = int64(rng.Intn(2*bound+1) - bound)
				if rng.Intn(4) == 0 {
					c[i][j] = int64(bound)
				} else if rng.Intn(4) == 0 {
					c[i][j] = int64(-bound)
				}
				edges = append(edges, ap.Edge[int8]{U: i, V: j, Cost: int8(c[i][j])})
			}
		}
		a, err := lsap.NewSparse(n, edges)
		if err != nil {
			t.Fatalf("%v: NewSparse: %v", c, err)
		}
		for step := 0; ; step++ {
			// Every other step removes an assigned edge, so a correct solver
			// becomes infeasible within 2n² steps.
			if step > 2*n*n {
				t.Fatalf("%v: still feasible after %v steps", c, step)
			}
			want, feasible := bruteforce.MinSum(c, mask)
			p, err := a.Assign()
			if !feasible {
				if !errors.Is(err, ap.ErrInfeasible) {
					t.Errorf("%v mask %v: got (%v, %v); want ErrInfeasible", c, mask, p, err)
				}
				break
			}
			if err != nil || int64(a.Cost()) != want {
				t.Fatalf("%v mask %v: got (%v, %v, cost %v); want cost %v", c, mask, p, err, a.Cost(), want)
			}
			u := rng.Intn(n)
			v := p[u]
			if step%2 == 1 {
				v = rng.Intn(n)
			}
			mask[u][v] = false
			a.Remove(u, v)
		}
	}
}
