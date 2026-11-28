package lbap_test

import (
	"errors"
	"math/rand/v2"
	"testing"

	"github.com/ryanjoneil/ap"
	"github.com/ryanjoneil/ap/lbap"
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

type fix struct {
	u, v  int
	value bool
}

func TestFixMatchesBruteForce(t *testing.T) {
	rng := rand.New(rand.NewPCG(3, 3))
	for range 1000 {
		n := 1 + rng.IntN(6)
		c := randomCosts(rng, n)
		if rng.IntN(2) == 0 {
			for i := range c { // few distinct costs, so many ties
				for j := range c[i] {
					c[i][j] = rng.Int64N(3)
				}
			}
		}
		smask := fullMask(n)
		for i := range smask {
			for j := range smask[i] {
				smask[i][j] = rng.Float64() < 0.6
			}
		}
		d, err := lbap.New(c)
		if err != nil {
			t.Fatal(err)
		}
		s, err := lbap.NewSparse(n, maskedEdges(c, smask))
		if err != nil {
			t.Fatal(err)
		}
		for name, k := range map[string]struct {
			a    *lbap.LBAP[int64]
			mask [][]bool
		}{"dense": {d, fullMask(n)}, "sparse": {s, smask}} {
			var p ap.Permutation
			if rng.IntN(2) == 0 {
				p = check(t, name, k.a, c, k.mask)
			}
			for range 3 {
				for f := 1 + rng.IntN(3); f > 0; f-- {
					u, v := rng.IntN(n), rng.IntN(n)
					if p != nil && rng.IntN(2) == 0 {
						v = p[u]
					}
					value := rng.IntN(2) == 0
					fixMask(k.mask, u, v, value)
					k.a.Fix(u, v, value)
				}
				p = check(t, name, k.a, c, k.mask)
			}
		}
	}
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
			a, err := lbap.New(c)
			if err != nil {
				t.Fatal(err)
			}
			if solveFirst {
				if _, err := a.Assign(); err != nil {
					t.Fatalf("%s: Assign: %v", name, err)
				}
			}
			mask := fullMask(len(c))
			for _, f := range tc.fixes {
				if f.u >= 0 && f.u < len(c) && f.v >= 0 && f.v < len(c) {
					fixMask(mask, f.u, f.v, f.value)
				}
				a.Fix(f.u, f.v, f.value)
			}
			p, err := a.Assign()
			if tc.infeasible != errors.Is(err, ap.ErrInfeasible) {
				t.Errorf("%s (solve first %v): Assign = (%v, %v); want infeasible %v",
					name, solveFirst, p, err, tc.infeasible)
				continue
			}
			if tc.infeasible {
				// Conflicts are sticky.
				if _, err := a.Assign(); !errors.Is(err, ap.ErrInfeasible) {
					t.Errorf("%s: second Assign = %v; want ErrInfeasible", name, err)
				}
				continue
			}
			check(t, name, a, c, mask)
		}
	}

	// Requiring a missing edge of a sparse solver is a conflict, and
	// forbidding one is a no-op.
	edges := []ap.Edge[int64]{{U: 0, V: 0, Cost: 1}, {U: 1, V: 1, Cost: 1}, {U: 0, V: 1, Cost: 1}}
	a, err := lbap.NewSparse(2, edges)
	if err != nil {
		t.Fatal(err)
	}
	b := a.Copy()
	b.Fix(1, 0, false)
	if _, err := b.Assign(); err != nil || b.Cost() != 1 {
		t.Errorf("forbid missing: got %v, cost %v; want cost 1", err, b.Cost())
	}
	a.Fix(1, 0, true)
	if p, err := a.Assign(); !errors.Is(err, ap.ErrInfeasible) {
		t.Errorf("require missing: Assign = (%v, %v); want ErrInfeasible", p, err)
	}
}

// Requirements made before Copy reach both solvers, and later ones stay on
// their own.
func TestFixCopyIndependent(t *testing.T) {
	rng := rand.New(rand.NewPCG(4, 4))
	for range 300 {
		n := 3 + rng.IntN(4)
		c := randomCosts(rng, n)
		a, err := lbap.New(c)
		if err != nil {
			t.Fatal(err)
		}
		mask := fullMask(n)
		u, v := rng.IntN(n), rng.IntN(n)
		fixMask(mask, u, v, true)
		a.Fix(u, v, true)

		b, bMask := a.Copy(), fullMask(n)
		fixMask(bMask, u, v, true)
		for range 2 {
			u, v = rng.IntN(n), rng.IntN(n)
			fixMask(bMask, u, v, true)
			b.Fix(u, v, true)
			u, v = rng.IntN(n), rng.IntN(n)
			fixMask(mask, u, v, true)
			a.Fix(u, v, true)
		}
		check(t, "original", a, c, mask)
		check(t, "copy", b, c, bMask)
	}
}
