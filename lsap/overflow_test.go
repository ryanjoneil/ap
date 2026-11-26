package lsap_test

import (
	"errors"
	"math/rand/v2"
	"testing"

	"github.com/ryanjoneil/ap"
	"github.com/ryanjoneil/ap/internal/bruteforce"
	"github.com/ryanjoneil/ap/internal/instances"
	"github.com/ryanjoneil/ap/internal/intmath"
	"github.com/ryanjoneil/ap/lsap"
)

// toT converts c to T. The caller ensures every cost fits.
func toT[T ap.Integer](c [][]int64) [][]T {
	ct := make([][]T, len(c))
	for i, row := range c {
		ct[i] = make([]T, len(row))
		for j, cij := range row {
			ct[i][j] = T(cij)
		}
	}
	return ct
}

// buildDense builds a dense solver for c. Every edge exists, so it leaves
// mask all true.
func buildDense[T ap.Integer](c [][]int64, _ [][]bool) (*lsap.LSAP[T], error) {
	return lsap.New(toT[T](c))
}

// denseWantErr reports whether New must reject c for T: some cost is the
// smallest value of T, or M = n·(maxC − minC) + maxC + 1 doesn't fit below
// the largest value of T. The costs are small enough that M fits in int64.
func denseWantErr[T ap.Integer](c [][]int64, _ [][]bool) bool {
	minC, maxC := c[0][0], c[0][0]
	for _, row := range c {
		for _, cij := range row {
			minC = min(minC, cij)
			maxC = max(maxC, cij)
		}
	}
	m := int64(len(c))*(maxC-minC) + maxC + 1
	return minC == int64(intmath.MinOf[T]()) || m >= int64(intmath.MaxOf[T]())
}

// A single cost c gives A = c and v = c, so the dense margin
// B = (127 − c − 1)/2 holds v exactly when c ≤ 42.
func TestDenseMarginInt8(t *testing.T) {
	a, err := lsap.New([][]int8{{42}})
	if err != nil {
		t.Fatalf("New(42): %v", err)
	}
	if p, err := a.Assign(); err != nil || a.Cost() != 42 {
		t.Errorf("Assign(42) = (%v, %v), cost %v; want cost 42", p, err, a.Cost())
	}

	a, err = lsap.New([][]int8{{43}})
	if err != nil {
		t.Fatalf("New(43): %v", err)
	}
	if p, err := a.Assign(); !errors.Is(err, ap.ErrOverflow) {
		t.Errorf("Assign(43) = (%v, %v); want ErrOverflow", p, err)
	}
}

// Each of these passes New, then Assign returns ErrOverflow from a different
// check.
func TestDenseOverflowInt8(t *testing.T) {
	for name, c := range map[string][][]int8{
		// A = 127 leaves no margin at all.
		"no margin": {{-127}},
		// A = 42 and B = 42. Phase 1 sets v = [−42, −42], and phase 2 sets
		// u[1] = 1 + 42 = 43 > B.
		"phase 2": {{-42, -42}, {1, 1}},
		// A = 40 and B = 43, so the duals fit, but the cost 160 doesn't.
		"cost": {{40, 40, 40, 40}, {40, 40, 40, 40}, {40, 40, 40, 40}, {40, 40, 40, 40}},
	} {
		a, err := lsap.New(c)
		if err != nil {
			t.Fatalf("%s: New: %v", name, err)
		}
		if p, err := a.Assign(); !errors.Is(err, ap.ErrOverflow) {
			t.Errorf("%s: Assign = (%v, %v); want ErrOverflow", name, p, err)
		}
	}
}

func TestDenseNeverWrongInt8(t *testing.T) {
	checkNeverWrong[int8](t, buildDense[int8], denseWantErr[int8])
}

func TestDenseNeverWrongInt16(t *testing.T) {
	checkNeverWrong[int16](t, buildDense[int16], denseWantErr[int16])
}

// buildSparse returns a build function for checkNeverWrong that makes an edge
// of each entry of c, drops a random quarter of them as missing edges, and
// builds a sparse solver from the rest.
func buildSparse[T ap.Integer](rng *rand.Rand) func(c [][]int64, mask [][]bool) (*lsap.LSAP[T], error) {
	return func(c [][]int64, mask [][]bool) (*lsap.LSAP[T], error) {
		var edges []ap.Edge[T]
		for i, row := range c {
			for j, cij := range row {
				if rng.IntN(4) == 0 {
					mask[i][j] = false
					continue
				}
				edges = append(edges, ap.Edge[T]{U: i, V: j, Cost: T(cij)})
			}
		}
		return lsap.NewSparse(len(c), edges)
	}
}

// sparseWantErr reports whether NewSparse must reject the edges of c left in
// mask for T: some cost is the smallest or largest value of T.
func sparseWantErr[T ap.Integer](c [][]int64, mask [][]bool) bool {
	for i, row := range c {
		for j, cij := range row {
			if mask[i][j] && (cij == int64(intmath.MinOf[T]()) || cij == int64(intmath.MaxOf[T]())) {
				return true
			}
		}
	}
	return false
}

// A single cost c gives A = c and v = c, so the sparse margin
// B = (127 − c − 1)/4 holds v exactly when c ≤ 25.
func TestSparseMarginInt8(t *testing.T) {
	a, err := lsap.NewSparse(1, []ap.Edge[int8]{{U: 0, V: 0, Cost: 25}})
	if err != nil {
		t.Fatalf("NewSparse(25): %v", err)
	}
	if p, err := a.Assign(); err != nil || a.Cost() != 25 {
		t.Errorf("Assign(25) = (%v, %v), cost %v; want cost 25", p, err, a.Cost())
	}

	a, err = lsap.NewSparse(1, []ap.Edge[int8]{{U: 0, V: 0, Cost: 26}})
	if err != nil {
		t.Fatalf("NewSparse(26): %v", err)
	}
	if p, err := a.Assign(); !errors.Is(err, ap.ErrOverflow) {
		t.Errorf("Assign(26) = (%v, %v); want ErrOverflow", p, err)
	}
}

// The largest accepted cost leaves no margin, and its column is not mistaken
// for an empty one.
func TestSparseMaxCostInt8(t *testing.T) {
	a, err := lsap.NewSparse(1, []ap.Edge[int8]{{U: 0, V: 0, Cost: 126}})
	if err != nil {
		t.Fatalf("NewSparse(126): %v", err)
	}
	if p, err := a.Assign(); !errors.Is(err, ap.ErrOverflow) {
		t.Errorf("Assign(126) = (%v, %v); want ErrOverflow", p, err)
	}
}

// Each of these passes NewSparse, then Assign returns ErrOverflow from a
// different check. Where there are removals, Assign must succeed before each
// one, and return ErrOverflow after the last.
func TestSparseOverflowInt8(t *testing.T) {
	for name, tc := range map[string]struct {
		n       int
		edges   []ap.Edge[int8]
		removes [][2]int
	}{
		// A = 127 leaves no margin at all.
		"no margin": {1, []ap.Edge[int8]{{U: 0, V: 0, Cost: -127}}, nil},
		// A = 25 and B = 25. Column reduction sets v = [−25, −25], and row
		// reduction sets u[1] = 25 + 25 = 50 > B.
		"initialize u": {2, []ap.Edge[int8]{
			{U: 0, V: 0, Cost: -25}, {U: 0, V: 1, Cost: -25},
			{U: 1, V: 0, Cost: 25}, {U: 1, V: 1, Cost: 25},
		}, nil},
		// A = 47, B = 19 and L = 41. Initializing gives u = 0, v = [−19, −17]
		// and the diagonal. After removing (0, 0), the search from row 0
		// labels column 1 with 61 > L. Scanning it anyway would label column
		// 0 with 61 + 47 + 19 = 127, the value that marks an unlabelled
		// column, so the search would miss column 0 and wrongly report no
		// assignment.
		"label cap": {2, []ap.Edge[int8]{
			{U: 0, V: 0, Cost: -19}, {U: 0, V: 1, Cost: 44},
			{U: 1, V: 0, Cost: 47}, {U: 1, V: 1, Cost: -17},
		}, [][2]int{{0, 0}}},
		// A = 24 and B = 25. Initializing gives u = [9, 0] and the diagonal.
		// After removing (0, 0), the shortest path from row 0 has length 21,
		// so u[0] would become 30 > B.
		"dual u[s]": {2, []ap.Edge[int8]{
			{U: 0, V: 0, Cost: 24}, {U: 0, V: 1, Cost: 13},
			{U: 1, V: 0, Cost: 15}, {U: 1, V: 1, Cost: -17},
		}, [][2]int{{0, 0}}},
		// A = 17 and B = 27. Initializing gives u = [15, 0] and the diagonal.
		// After removing (1, 1), the shortest path from row 1 has length 19
		// and scans column 0 at label 0, so u[0] would become 34 > B.
		"dual u[i]": {2, []ap.Edge[int8]{
			{U: 0, V: 0, Cost: 16}, {U: 0, V: 1, Cost: 17},
			{U: 1, V: 0, Cost: 1}, {U: 1, V: 1, Cost: -17},
		}, [][2]int{{1, 1}}},
		// A = 20 and B = 26. Initializing gives v = [−20, −19] and the
		// diagonal. After removing (1, 1), the shortest path from row 1 has
		// length 24 and scans column 0 at label 7, so v[0] would become
		// −20 − 17 = −37 < −B.
		"dual v[j]": {2, []ap.Edge[int8]{
			{U: 0, V: 0, Cost: -20}, {U: 0, V: 1, Cost: -2},
			{U: 1, V: 0, Cost: -13}, {U: 1, V: 1, Cost: -19},
		}, [][2]int{{1, 1}}},
		// A = 38 and B = 22. Initializing gives v = 20 and u = [0, 18, 18, 18]
		// and assigns every row, so the duals fit, but the cost
		// 20 + 3·38 = 134 doesn't.
		"cost": {4, func() []ap.Edge[int8] {
			var edges []ap.Edge[int8]
			for i := 0; i < 4; i++ {
				for j := 0; j < 4; j++ {
					cost := int8(38)
					if i == 0 {
						cost = 20
					}
					edges = append(edges, ap.Edge[int8]{U: i, V: j, Cost: cost})
				}
			}
			return edges
		}(), nil},
	} {
		a, err := lsap.NewSparse(tc.n, tc.edges)
		if err != nil {
			t.Fatalf("%s: NewSparse: %v", name, err)
		}
		for _, r := range tc.removes {
			if p, err := a.Assign(); err != nil {
				t.Fatalf("%s: Assign before Remove(%v, %v) = (%v, %v)", name, r[0], r[1], p, err)
			}
			a.Remove(r[0], r[1])
		}
		if p, err := a.Assign(); !errors.Is(err, ap.ErrOverflow) {
			t.Errorf("%s: Assign = (%v, %v), cost %v; want ErrOverflow", name, p, err, a.Cost())
		}
	}
}

func TestSparseNeverWrongInt8(t *testing.T) {
	rng := rand.New(rand.NewPCG(3, 4))
	checkNeverWrong[int8](t, buildSparse[int8](rng), sparseWantErr[int8])
}

func TestSparseNeverWrongInt16(t *testing.T) {
	rng := rand.New(rand.NewPCG(3, 4))
	checkNeverWrong[int16](t, buildSparse[int16](rng), sparseWantErr[int16])
}

// checkNeverWrong runs 2000 random trials of at most 4×4 problems with costs
// in [−K, K], where K ranges from 1 to the largest value of T less one. Each
// trial builds a solver with build, then alternately removes an assigned edge
// and a random edge, solving after each removal, until brute force (in int64)
// says the problem is infeasible. Every Assign must either return
// ap.ErrOverflow, which ends the trial, or agree with brute force.
//
// checkNeverWrong passes build an all-true mask, and build sets mask[i][j] to
// false for each edge (i, j) it leaves out of the solver. wantErr reports
// whether build must fail for c with the edges left in mask; trials for which
// it does are skipped after checking that build failed.
//
// Over all trials, at least one Assign must solve and at least one must
// return ap.ErrOverflow.
func checkNeverWrong[T ap.Integer](
	t *testing.T,
	build func(c [][]int64, mask [][]bool) (*lsap.LSAP[T], error),
	wantErr func(c [][]int64, mask [][]bool) bool,
) {
	t.Helper()

	maxT := int64(intmath.MaxOf[T]())
	ks := []int64{1, 2, 5, 10, 20, 40, maxT / 4, maxT / 2, maxT - 1}
	rng := rand.New(rand.NewPCG(1, 2))
	solved, overflowed := 0, 0

	for trial := 0; trial < 2000; trial++ {
		n := 1 + rng.IntN(4)
		k := ks[rng.IntN(len(ks))]
		c := make([][]int64, n)
		mask := make([][]bool, n)
		for i := range c {
			c[i] = make([]int64, n)
			mask[i] = make([]bool, n)
			for j := range c[i] {
				c[i][j] = rng.Int64N(2*k+1) - k
				mask[i][j] = true
			}
		}

		a, err := build(c, mask)
		if want := wantErr(c, mask); want != (err != nil) {
			t.Fatalf("%v: build error = %v; want error %v", c, err, want)
		}
		if err != nil {
			continue
		}

		for step := 0; ; step++ {
			// Every other step removes an assigned edge, so a correct solver
			// becomes infeasible within 2n² steps.
			if step > 2*n*n {
				t.Fatalf("%v: still feasible after %v steps", c, step)
			}
			want, feasible := bruteforce.MinSum(c, mask)
			p, err := a.Assign()
			if errors.Is(err, ap.ErrOverflow) {
				overflowed++
				// The overflow is sticky, and carried by Copy.
				if _, err := a.Assign(); !errors.Is(err, ap.ErrOverflow) {
					t.Fatalf("%v mask %v: second Assign = %v; want ErrOverflow", c, mask, err)
				}
				if _, err := a.Copy().Assign(); !errors.Is(err, ap.ErrOverflow) {
					t.Fatalf("%v mask %v: Copy().Assign() = %v; want ErrOverflow", c, mask, err)
				}
				break
			}
			if !feasible {
				if !errors.Is(err, ap.ErrInfeasible) {
					t.Fatalf("%v mask %v: got (%v, %v); want ErrInfeasible or ErrOverflow", c, mask, p, err)
				}
				break
			}
			if err != nil {
				t.Fatalf("%v mask %v: Assign: %v", c, mask, err)
			}
			solved++
			if got := int64(a.Cost()); got != want {
				t.Fatalf("%v mask %v: Cost() = %v; want %v", c, mask, got, want)
			}
			if got := permutationCost(c, p); got != want {
				t.Fatalf("%v mask %v: permutation %v costs %v; want %v", c, mask, p, got, want)
			}
			for u, v := range p {
				if !mask[u][v] {
					t.Fatalf("%v mask %v: %v uses removed edge (%v, %v)", c, mask, p, u, v)
				}
			}

			u := rng.IntN(n)
			v := p[u]
			if step%2 == 1 {
				v = rng.IntN(n)
			}
			mask[u][v] = false
			a.Remove(u, v)
		}
	}

	if solved == 0 || overflowed == 0 {
		t.Errorf("%v steps solved and %v overflowed; want both > 0", solved, overflowed)
	}
}

// toEdgesT converts edges to T. The caller ensures every cost fits.
func toEdgesT[T ap.Integer](edges []ap.Edge[int64]) []ap.Edge[T] {
	et := make([]ap.Edge[T], len(edges))
	for k, e := range edges {
		et[k] = ap.Edge[T]{U: e.U, V: e.V, Cost: T(e.Cost)}
	}
	return et
}

// The benchmark classes solve in int32 at their benchmark sizes, with the same
// cost as in int64, both dense and sparse. The sparse instances have the
// book's degree.
func TestNoFalseAlarmsInt32(t *testing.T) {
	if testing.Short() {
		t.Skip("large instances")
	}

	for name, tc := range map[string]struct {
		n     int
		class func() instances.Class // A fresh class, seeded the same each call
	}{
		"uniform": {1000, func() instances.Class {
			return instances.Uniform(rand.New(rand.NewPCG(1, 1)), 1_000_000)
		}},
		"geometric": {1000, func() instances.Class {
			return instances.Geometric(rand.New(rand.NewPCG(1, 2)), 1000, 1_000_000)
		}},
		"two-cost": {1000, func() instances.Class {
			return instances.TwoCost(rand.New(rand.NewPCG(1, 3)))
		}},
		"machol-wien": {300, instances.MacholWien},
	} {
		t.Run(name+"/dense", func(t *testing.T) {
			c := instances.Dense(tc.n, tc.class())
			checkSameCost(t, func() (*lsap.LSAP[int64], error) { return lsap.New(c) },
				func() (*lsap.LSAP[int32], error) { return lsap.New(toT[int32](c)) })
		})
		t.Run(name+"/sparse", func(t *testing.T) {
			rng := rand.New(rand.NewPCG(2, 1))
			edges := instances.Sparse(rng, tc.n, instances.BookDegree(tc.n), tc.class())
			checkSameCost(t, func() (*lsap.LSAP[int64], error) { return lsap.NewSparse(tc.n, edges) },
				func() (*lsap.LSAP[int32], error) { return lsap.NewSparse(tc.n, toEdgesT[int32](edges)) })
		})
	}
}

// checkSameCost builds and solves a problem in int64 and in int32, and checks that
// both solve with the same cost.
func checkSameCost(
	t *testing.T,
	build64 func() (*lsap.LSAP[int64], error),
	build32 func() (*lsap.LSAP[int32], error),
) {
	t.Helper()

	a64, err := build64()
	if err != nil {
		t.Fatalf("build int64: %v", err)
	}
	if _, err := a64.Assign(); err != nil {
		t.Fatalf("Assign int64: %v", err)
	}

	a32, err := build32()
	if err != nil {
		t.Fatalf("build int32: %v", err)
	}
	if _, err := a32.Assign(); err != nil {
		t.Fatalf("Assign int32: %v", err)
	}
	if got, want := int64(a32.Cost()), a64.Cost(); got != want {
		t.Errorf("int32 cost %v; int64 cost %v", got, want)
	}
}
