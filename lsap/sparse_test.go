package lsap_test

import (
	"errors"
	"math"
	"math/rand"
	"reflect"
	"strconv"
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
func TestSparseForbidWholeRow(t *testing.T) {
	a, _ := lsap.NewSparse(2, []ap.Edge[int64]{
		{U: 0, V: 0, Cost: 1}, {U: 0, V: 1, Cost: 2}, {U: 1, V: 0, Cost: 3}, {U: 1, V: 1, Cost: 4},
	})
	_, _ = a.Assign()
	b := a.Copy()
	a.Fix(0, 0, false)
	a.Fix(0, 1, false)
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
	a.Fix(0, 0, false) // Not assigned; the copy must inherit this removal.
	b := a.Copy()
	b.Fix(0, 1, false)
	b.Fix(1, 1, false)
	pa, _ := a.Assign()
	if a.Cost() != 5 || !reflect.DeepEqual(pa, ap.Permutation{1, 0, 2}) {
		t.Errorf("original changed: %v %v", pa, a.Cost())
	}
	if _, ok := a.ReducedCost(1, 1); !ok {
		t.Error("Fix on the copy forbade the edge in the original")
	}
	a.Fix(2, 2, false)
	if _, ok := b.ReducedCost(2, 2); !ok {
		t.Error("Fix on the original forbade the edge in the copy")
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

// Forbid each assigned edge in turn, re-solve, and compare with brute force on
// the shrinking mask.
func TestSparseForbidIncremental(t *testing.T) {
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
	} {
		if _, err := lsap.NewSparse(tc.n, tc.edges); err == nil {
			t.Errorf("%s: want error", name)
		}
	}
	// n = math.MaxInt32 + 1 only fits in a 64-bit int. Computing it at run
	// time keeps the test compiling on 32-bit platforms.
	if strconv.IntSize == 64 {
		n := math.MaxInt32
		n++
		if _, err := lsap.NewSparse[int64](n, nil); err == nil {
			t.Errorf("n = %d: want error", n)
		}
	}
	// NewSparse rejects only costs equal to the smallest or largest value of
	// T, whatever n is.
	for _, c := range []int8{math.MaxInt8, math.MinInt8} {
		if _, err := lsap.NewSparse(2, []ap.Edge[int8]{{U: 0, V: 0, Cost: c}}); err == nil {
			t.Errorf("int8 cost %d: want error", c)
		}
	}
	for _, c := range []int8{11, -11} {
		if _, err := lsap.NewSparse(2, []ap.Edge[int8]{{U: 0, V: 0, Cost: c}}); err != nil {
			t.Errorf("int8 cost %d: %v", c, err)
		}
	}
}

func TestNewSparseCopiesInput(t *testing.T) {
	edges := []ap.Edge[int64]{{U: 1, V: 1, Cost: 4}, {U: 0, V: 0, Cost: 1}, {U: 0, V: 1, Cost: 2}}
	orig := append([]ap.Edge[int64](nil), edges...)
	a := newSparse(t, 2, edges)
	a.Fix(0, 0, false)
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
	a.Fix(0, 1, false) // missing: a no-op
	a.Fix(5, 5, false) // out of range: a no-op
	if _, err := a.Assign(); err != nil || a.Cost() != 2 {
		t.Errorf("got %v, cost %v", err, a.Cost())
	}
}
