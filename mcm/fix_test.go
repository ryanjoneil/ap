package mcm_test

import (
	"errors"
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/ryanjoneil/ap"
	"github.com/ryanjoneil/ap/internal/bruteforce"
	"github.com/ryanjoneil/ap/mcm"
)

// fixMask applies Fix(u, v, value) to the brute-force mask: true clears every
// other entry of row u and column v, and false clears (u, v). The mask may be
// rectangular. Applying the fixes in order gives the expected result.
func fixMask(mask [][]bool, u, v int, value bool) {
	if !value {
		mask[u][v] = false
		return
	}
	for k := range mask[u] {
		if k != v {
			mask[u][k] = false
		}
	}
	for k := range mask {
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
	for range 2000 {
		nU, nV := 1+rng.IntN(5), 1+rng.IntN(6)
		if rng.IntN(2) == 0 {
			nU, nV = nV, nU
		}
		ok := randOK(rng, nU, nV, 0.6)
		m, err := mcm.NewSparse(nU, nV, pairsOf(ok))
		if err != nil {
			continue // no edges
		}
		required := map[int]int{}
		var mate []int
		if rng.IntN(2) == 0 {
			mate = match(t, m)
		}
		for range 3 {
			for f := 1 + rng.IntN(3); f > 0; f-- {
				u, v := rng.IntN(nU), rng.IntN(nV)
				if mate != nil && mate[u] >= 0 && rng.IntN(2) == 0 {
					v = mate[u]
				}
				value := rng.IntN(2) == 0
				// A required edge must be allowed and not blocked by another
				// requirement; skip conflicting fixes here, which
				// TestFixConflicts covers.
				if value {
					blocked := !ok[u][v]
					for ru, rv := range required {
						if (ru == u) != (rv == v) {
							blocked = true
						}
					}
					if blocked {
						continue
					}
					required[u] = v
				} else if rv, has := required[u]; has && rv == v {
					continue
				}
				fixMask(ok, u, v, value)
				m.Fix(u, v, value)
			}
			mate = match(t, m)
			n := checkValid(t, mate, ok)
			if want := bruteforce.MaxMatching(ok); m.Size() != want || n != want {
				t.Fatalf("mask %v: size %d (pairs %d); want %d", ok, m.Size(), n, want)
			}
			for u, v := range required {
				if mate[u] != v {
					t.Fatalf("mask %v: mate %v lacks required edge (%d, %d)", ok, mate, u, v)
				}
			}
		}
	}
}

func TestFixConflicts(t *testing.T) {
	adj := [][]bool{{true, true, true}, {true, true, true}, {true, true, true}}
	for name, tc := range map[string]struct {
		fixes      []fix
		infeasible bool
	}{
		"same row":        {[]fix{{0, 0, true}, {0, 1, true}}, true},
		"same column":     {[]fix{{0, 0, true}, {1, 0, true}}, true},
		"forbid, require": {[]fix{{0, 0, false}, {0, 0, true}}, true},
		"require, forbid": {[]fix{{0, 0, true}, {0, 0, false}}, true},
		"require twice":   {[]fix{{0, 0, true}, {0, 0, true}}, false},
		"forbid twice":    {[]fix{{0, 0, false}, {0, 0, false}}, false},
		"disjoint":        {[]fix{{0, 0, true}, {1, 1, true}}, false},
		"forbid row-mate": {[]fix{{0, 0, true}, {0, 1, false}}, false},
		"out of range":    {[]fix{{0, 0, true}, {3, 0, true}, {0, -1, true}, {0, 3, false}}, false},
	} {
		for _, solveFirst := range []bool{false, true} {
			m, err := mcm.New(adj)
			if err != nil {
				t.Fatal(err)
			}
			if solveFirst {
				match(t, m)
			}
			for _, f := range tc.fixes {
				m.Fix(f.u, f.v, f.value)
			}
			mate, err := m.Match()
			p, perr := m.Assign()
			if !tc.infeasible {
				if err != nil || perr != nil {
					t.Errorf("%s (solve first %v): Match = (%v, %v), Assign = (%v, %v); want no error",
						name, solveFirst, mate, err, p, perr)
				}
				continue
			}
			if mate != nil || !errors.Is(err, ap.ErrInfeasible) || m.Size() != 0 ||
				p != nil || !errors.Is(perr, ap.ErrInfeasible) {
				t.Errorf("%s (solve first %v): Match = (%v, %v), Size %d, Assign = (%v, %v); want (nil, ErrInfeasible), 0, ErrInfeasible",
					name, solveFirst, mate, err, m.Size(), p, perr)
			}
			// Conflicts are sticky.
			if _, err := m.Match(); !errors.Is(err, ap.ErrInfeasible) {
				t.Errorf("%s: second Match = %v; want ErrInfeasible", name, err)
			}
		}
	}

	// Requiring a missing edge is a conflict, and forbidding one is a no-op.
	m, err := mcm.NewSparse(2, 2, []ap.Pair{{U: 0, V: 0}, {U: 1, V: 1}})
	if err != nil {
		t.Fatal(err)
	}
	b := m.Copy()
	b.Fix(0, 1, false)
	if mate := match(t, b); !slices.Equal(mate, []int{0, 1}) {
		t.Errorf("forbid missing: got %v", mate)
	}
	m.Fix(0, 1, true)
	if mate, err := m.Match(); mate != nil || !errors.Is(err, ap.ErrInfeasible) {
		t.Errorf("require missing: Match = (%v, %v); want (nil, ErrInfeasible)", mate, err)
	}
}
