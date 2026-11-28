package lsap_test

import (
	"errors"
	"math/rand"
	"slices"
	"testing"

	"github.com/ryanjoneil/ap"
	"github.com/ryanjoneil/ap/lsap"
)

// fixMask applies Fix(u, v, value) to the brute-force mask: true clears every
// other entry of row u and column v, and false clears (u, v). Applying the
// fixes in order gives the expected result, including every conflict, which
// leaves a row with no allowed entry.
func fixMask(mask [][]bool, u, v int, value bool) {
	if !value {
		mask[u][v] = false
		return
	}
	for k := range mask {
		if k != v {
			mask[u][k] = false
		}
		if k != u {
			mask[k][v] = false
		}
	}
}

// maskedKernel is a solver and the mask of edges it allows.
type maskedKernel struct {
	name string
	a    *lsap.LSAP[int64]
	mask [][]bool
}

// fixKernels returns both kernels on the full matrix c, and a sparse solver
// on a random subset of about 60% of its edges.
func fixKernels(t *testing.T, rng *rand.Rand, c [][]int64) []maskedKernel {
	t.Helper()

	var ks []maskedKernel
	for _, k := range bothKernels(t, c) {
		ks = append(ks, maskedKernel{k.name, k.a, fullMask(len(c))})
	}

	mask := fullMask(len(c))
	var edges []ap.Edge[int64]
	for i, row := range c {
		for j, cij := range row {
			if rng.Float64() >= 0.6 {
				mask[i][j] = false
				continue
			}
			edges = append(edges, ap.Edge[int64]{U: i, V: j, Cost: cij})
		}
	}
	s, err := lsap.NewSparse(len(c), edges)
	if err != nil {
		t.Fatalf("%v: NewSparse: %v", edges, err)
	}
	return append(ks, maskedKernel{"sparse 60%", s, mask})
}

func cloneMask(mask [][]bool) [][]bool {
	out := make([][]bool, len(mask))
	for i, row := range mask {
		out[i] = slices.Clone(row)
	}
	return out
}

func TestFixMatchesBruteForce(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	for trial := 0; trial < 500; trial++ {
		n := 1 + rng.Intn(6)
		c := randomCosts(rng, n)
		for _, k := range fixKernels(t, rng, c) {
			p := checkAssignMask(t, k.name, k.a, c, k.mask)
			for round := 0; round < 3; round++ {
				for f := 1 + rng.Intn(4); f > 0; f-- {
					u, v := rng.Intn(n), rng.Intn(n)
					if p != nil && rng.Intn(2) == 0 {
						v = p[u]
					}
					value := rng.Intn(2) == 0
					fixMask(k.mask, u, v, value)
					k.a.Fix(u, v, value)
				}
				if p = checkAssignMask(t, k.name, k.a, c, k.mask); p != nil {
					checkDuals(t, k.name, k.a, c, k.mask, p)
				}
			}
		}
	}
}

type fix struct {
	u, v  int
	value bool
}

func TestFixConflicts(t *testing.T) {
	c := [][]int64{{4, 1, 3}, {2, 0, 5}, {3, 2, 2}}
	for name, tc := range map[string]struct {
		fixes      []fix
		infeasible bool
	}{
		"same row":          {[]fix{{0, 0, true}, {0, 1, true}}, true},
		"same column":       {[]fix{{0, 0, true}, {1, 0, true}}, true},
		"forbid, require":   {[]fix{{0, 0, false}, {0, 0, true}}, true},
		"require, forbid":   {[]fix{{0, 0, true}, {0, 0, false}}, true},
		"require twice":     {[]fix{{0, 0, true}, {0, 0, true}}, false},
		"forbid twice":      {[]fix{{0, 0, false}, {0, 0, false}}, false},
		"disjoint":          {[]fix{{0, 0, true}, {1, 1, true}}, false},
		"forbid row-mate":   {[]fix{{0, 0, true}, {0, 1, false}}, false},
		"out of range":      {[]fix{{0, 0, true}, {3, 0, true}, {0, -1, true}, {0, 3, false}}, false},
		"leaves infeasible": {[]fix{{1, 1, false}, {1, 2, false}, {0, 0, true}}, true},
	} {
		for _, solveFirst := range []bool{false, true} {
			for _, k := range bothKernels(t, c) {
				if solveFirst {
					if _, err := k.a.Assign(); err != nil {
						t.Fatalf("%s: %s: Assign: %v", name, k.name, err)
					}
				}
				mask := fullMask(len(c))
				for _, f := range tc.fixes {
					if f.u >= 0 && f.u < len(c) && f.v >= 0 && f.v < len(c) {
						fixMask(mask, f.u, f.v, f.value)
					}
					k.a.Fix(f.u, f.v, f.value)
				}
				p, err := k.a.Assign()
				if tc.infeasible != errors.Is(err, ap.ErrInfeasible) {
					t.Errorf("%s: %s (solve first %v): Assign = (%v, %v); want infeasible %v",
						name, k.name, solveFirst, p, err, tc.infeasible)
					continue
				}
				if tc.infeasible {
					// Conflicts are sticky.
					if _, err := k.a.Assign(); !errors.Is(err, ap.ErrInfeasible) {
						t.Errorf("%s: %s: second Assign = %v; want ErrInfeasible", name, k.name, err)
					}
					continue
				}
				checkAssignMask(t, name+": "+k.name, k.a, c, mask)
			}
		}
	}

	// Requiring a missing edge of a sparse solver is a conflict, and
	// forbidding one is a no-op.
	edges := []ap.Edge[int64]{{U: 0, V: 0, Cost: 1}, {U: 1, V: 1, Cost: 1}, {U: 0, V: 1, Cost: 1}}
	a, err := lsap.NewSparse(2, edges)
	if err != nil {
		t.Fatal(err)
	}
	b := a.Copy()
	b.Fix(1, 0, false)
	if _, err := b.Assign(); err != nil || b.Cost() != 2 {
		t.Errorf("forbid missing: got %v, cost %v; want cost 2", err, b.Cost())
	}
	a.Fix(1, 0, true)
	if p, err := a.Assign(); !errors.Is(err, ap.ErrInfeasible) {
		t.Errorf("require missing: Assign = (%v, %v); want ErrInfeasible", p, err)
	}
}

// Review Focus 1: requiring an assigned edge keeps the solution, including
// the dual prices, so nothing is re-solved.
func TestFixAssignedKeepsSolution(t *testing.T) {
	rng := rand.New(rand.NewSource(5))
	for trial := 0; trial < 20; trial++ {
		c := randomCosts(rng, 5)
		for _, k := range bothKernels(t, c) {
			p0, err := k.a.Assign()
			if err != nil {
				t.Fatalf("%s: %v: Assign: %v", k.name, c, err)
			}
			cost0, duals0 := k.a.Cost(), k.a.DualPrices()
			for u := range p0 {
				k.a.Fix(u, p0[u], true)
				p, err := k.a.Assign()
				if err != nil || !slices.Equal(p, p0) || k.a.Cost() != cost0 {
					t.Fatalf("%s: %v: after requiring (%v, %v): got (%v, %v), cost %v; want %v, cost %v",
						k.name, c, u, p0[u], p, err, k.a.Cost(), p0, cost0)
				}
				d := k.a.DualPrices()
				if !slices.Equal(d.U, duals0.U) || !slices.Equal(d.V, duals0.V) {
					t.Fatalf("%s: %v: after requiring (%v, %v): duals %v; want %v",
						k.name, c, u, p0[u], d, duals0)
				}
			}
		}
	}
}

// Review Focus 2: requirements made before any Assign, and on a copy made
// before the first Assign, survive the first solve.
func TestFixBeforeAssign(t *testing.T) {
	rng := rand.New(rand.NewSource(2))
	for trial := 0; trial < 300; trial++ {
		n := 1 + rng.Intn(6)
		c := randomCosts(rng, n)
		for _, k := range fixKernels(t, rng, c) {
			b, bMask := k.a.Copy(), cloneMask(k.mask)
			for f := 1 + rng.Intn(3); f > 0; f-- {
				u, v := rng.Intn(n), rng.Intn(n)
				fixMask(k.mask, u, v, true)
				k.a.Fix(u, v, true)
				u, v = rng.Intn(n), rng.Intn(n)
				fixMask(bMask, u, v, true)
				b.Fix(u, v, true)
			}
			// A copy made after the requirements, before any Assign.
			cp := k.a.Copy()
			checkAssignMask(t, k.name+" fresh", k.a, c, k.mask)
			checkAssignMask(t, k.name+" copy before", b, c, bMask)
			checkAssignMask(t, k.name+" copy after", cp, c, k.mask)
		}
	}
}

// Review Focus 3: after requiring (u, v), its reduced cost is 0, and every
// other edge in row u or column v reads as forbidden.
func TestFixReducedCost(t *testing.T) {
	rng := rand.New(rand.NewSource(3))
	for trial := 0; trial < 300; trial++ {
		n := 2 + rng.Intn(5)
		c := randomCosts(rng, n)
		for _, k := range bothKernels(t, c) {
			p, err := k.a.Assign()
			if err != nil {
				t.Fatalf("%s: Assign: %v", k.name, err)
			}
			// Require an edge outside the assignment, so that the dual prices
			// move.
			u := rng.Intn(n)
			v := (p[u] + 1 + rng.Intn(n-1)) % n
			k.a.Fix(u, v, true)
			mask := fullMask(n)
			fixMask(mask, u, v, true)
			if checkAssignMask(t, k.name, k.a, c, mask) == nil {
				continue
			}
			if rc, ok := k.a.ReducedCost(u, v); rc != 0 || !ok {
				t.Errorf("%s: %v: ReducedCost(%v, %v) = (%v, %v); want (0, true)", k.name, c, u, v, rc, ok)
			}
			for w := 0; w < n; w++ {
				if w == v {
					continue
				}
				if rc, ok := k.a.ReducedCost(u, w); ok {
					t.Errorf("%s: %v: ReducedCost(%v, %v) = (%v, true); want false", k.name, c, u, w, rc)
				}
			}
			for x := 0; x < n; x++ {
				if x == u {
					continue
				}
				if rc, ok := k.a.ReducedCost(x, v); ok {
					t.Errorf("%s: %v: ReducedCost(%v, %v) = (%v, true); want false", k.name, c, x, v, rc)
				}
			}
		}
	}
}

// Requirements on a copy don't reach the original, and the reverse, while
// requirements made before Copy reach both.
func TestFixCopyIndependent(t *testing.T) {
	rng := rand.New(rand.NewSource(4))
	for trial := 0; trial < 300; trial++ {
		n := 3 + rng.Intn(4)
		c := randomCosts(rng, n)
		for _, k := range fixKernels(t, rng, c) {
			p := checkAssignMask(t, k.name, k.a, c, k.mask)
			if p == nil {
				continue
			}
			// Before Copy: require an assigned edge.
			u := rng.Intn(n)
			fixMask(k.mask, u, p[u], true)
			k.a.Fix(u, p[u], true)

			b, bMask := k.a.Copy(), cloneMask(k.mask)
			for f := 0; f < 2; f++ {
				u, v := rng.Intn(n), rng.Intn(n)
				fixMask(bMask, u, v, true)
				b.Fix(u, v, true)
				u, v = rng.Intn(n), rng.Intn(n)
				fixMask(k.mask, u, v, true)
				k.a.Fix(u, v, true)
			}
			checkAssignMask(t, k.name+" original", k.a, c, k.mask)
			checkAssignMask(t, k.name+" copy", b, c, bMask)
		}
	}
}

// Once Assign has returned ErrOverflow, a conflicting Fix doesn't change it.
func TestFixAfterOverflow(t *testing.T) {
	d, err := lsap.New([][]int8{{-42, -42}, {1, 1}})
	if err != nil {
		t.Fatal(err)
	}
	s, err := lsap.NewSparse(2, []ap.Edge[int8]{
		{U: 0, V: 0, Cost: -25}, {U: 0, V: 1, Cost: -25},
		{U: 1, V: 0, Cost: 25}, {U: 1, V: 1, Cost: 25},
	})
	if err != nil {
		t.Fatal(err)
	}
	for name, a := range map[string]*lsap.LSAP[int8]{"dense": d, "sparse": s} {
		if _, err := a.Assign(); !errors.Is(err, ap.ErrOverflow) {
			t.Fatalf("%s: Assign = %v; want ErrOverflow", name, err)
		}
		a.Fix(0, 0, true)
		a.Fix(1, 0, true)
		a.Fix(0, 0, false)
		if _, err := a.Assign(); !errors.Is(err, ap.ErrOverflow) {
			t.Errorf("%s: Assign after a conflict = %v; want ErrOverflow", name, err)
		}
		if _, err := a.Copy().Assign(); !errors.Is(err, ap.ErrOverflow) {
			t.Errorf("%s: Copy().Assign() after a conflict = %v; want ErrOverflow", name, err)
		}
	}
}
