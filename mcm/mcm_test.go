package mcm_test

import (
	"errors"
	"math/rand/v2"
	"slices"
	"strings"
	"testing"

	"github.com/ryanjoneil/ap"
	"github.com/ryanjoneil/ap/internal/bruteforce"
	"github.com/ryanjoneil/ap/mcm"
)

func randOK(rng *rand.Rand, nU, nV int, p float64) [][]bool {
	ok := make([][]bool, nU)
	for u := range ok {
		ok[u] = make([]bool, nV)
		for v := range ok[u] {
			ok[u][v] = rng.Float64() < p
		}
	}
	return ok
}

func pairsOf(ok [][]bool) []ap.Pair {
	var ps []ap.Pair
	for u, row := range ok {
		for v, b := range row {
			if b {
				ps = append(ps, ap.Pair{U: u, V: v})
			}
		}
	}
	return ps
}

func checkValid(t *testing.T, mate []int, ok [][]bool) int {
	t.Helper()
	seen := map[int]bool{}
	n := 0
	for u, v := range mate {
		if v < 0 {
			continue
		}
		if !ok[u][v] || seen[v] {
			t.Fatalf("invalid pair u=%d v=%d", u, v)
		}
		seen[v] = true
		n++
	}
	return n
}

func TestMatchBruteForce(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 1))
	for i := range 1000 {
		nU, nV := 1+rng.IntN(7), 1+rng.IntN(7)
		ok := randOK(rng, nU, nV, []float64{0.2, 0.5, 0.8}[i%3])
		m, err := mcm.NewSparse(nU, nV, pairsOf(ok))
		if err != nil && len(pairsOf(ok)) == 0 {
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		mate := m.Match()
		n := checkValid(t, mate, ok)
		if want := bruteforce.MaxMatching(ok); m.Size() != want || n != want {
			t.Fatalf("size %d (pairs %d) want %d", m.Size(), n, want)
		}
	}
}

func TestAssign(t *testing.T) {
	m, err := mcm.New([][]bool{{true, false}, {true, true}})
	if err != nil {
		t.Fatal(err)
	}
	p, err := m.Assign()
	if err != nil || !slices.Equal(p, ap.Permutation{0, 1}) {
		t.Fatalf("got %v, %v", p, err)
	}
	for _, adj := range [][][]bool{
		{{true, true, true}},
		{{true, false}, {true, false}},
	} {
		m, err := mcm.New(adj)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := m.Assign(); !errors.Is(err, ap.ErrInfeasible) {
			t.Fatalf("%v: got %v", adj, err)
		}
	}
	// Non-square: Assign still solves, so Size is valid.
	m, _ = mcm.New([][]bool{{true, true, true}})
	_, _ = m.Assign()
	if m.Size() != 1 {
		t.Fatalf("Size %d want 1", m.Size())
	}
}

func TestSizeBeforeMatch(t *testing.T) {
	m, _ := mcm.New([][]bool{{true}})
	if m.Size() != 0 {
		t.Fatalf("Size %d want 0", m.Size())
	}
}

func TestMatchReturnsCopy(t *testing.T) {
	m, _ := mcm.New([][]bool{{true}})
	a := m.Match()
	a[0] = 99
	if b := m.Match(); b[0] != 0 {
		t.Fatalf("Match not a copy: %v", b)
	}
}

// TestRemoveIncremental removes matched edges one at a time, then random
// edges (matched or not, repeated, out of range), checking against brute
// force each time. Removing a matched edge shrinks its row's window while
// the edge stays in Adj past end[u].
func TestRemoveIncremental(t *testing.T) {
	rng := rand.New(rand.NewPCG(2, 2))
	for range 300 {
		nU, nV := 1+rng.IntN(7), 1+rng.IntN(7)
		ok := randOK(rng, nU, nV, 0.6)
		if len(pairsOf(ok)) == 0 {
			continue
		}
		m, err := mcm.NewSparse(nU, nV, pairsOf(ok))
		if err != nil {
			t.Fatal(err)
		}
		for {
			mate := m.Match()
			checkValid(t, mate, ok)
			if m.Size() != bruteforce.MaxMatching(ok) {
				t.Fatalf("size %d want %d", m.Size(), bruteforce.MaxMatching(ok))
			}
			u := -1
			for i, v := range mate {
				if v >= 0 {
					u = i
					break
				}
			}
			if u < 0 {
				break
			}
			v := mate[u]
			m.Remove(u, v)
			ok[u][v] = false
			m.Remove(u, v) // no-op
			m.Remove(-1, 0)
			m.Remove(0, nV)
			m.Remove(nU, 0)
			if got := m.Match(); got[u] == v {
				t.Fatalf("Match pairs %d with removed %d", u, v)
			}
			// Remove an arbitrary other edge, matched or not.
			ru, rv := rng.IntN(nU), rng.IntN(nV)
			m.Remove(ru, rv)
			ok[ru][rv] = false
			if got := m.Match(); got[ru] == rv {
				t.Fatalf("Match pairs %d with removed %d", ru, rv)
			}
		}
	}
}

func TestRemoveUnmatchedKeepsSolved(t *testing.T) {
	m, _ := mcm.New([][]bool{{true, true}, {false, true}})
	if got := m.Match(); !slices.Equal(got, []int{0, 1}) {
		t.Fatalf("got %v", got)
	}
	m.Remove(0, 1)
	if m.Size() != 2 {
		t.Fatalf("Size %d want 2", m.Size())
	}
}

func TestCopyIndependent(t *testing.T) {
	ok := [][]bool{{true, true, false}, {true, true, true}, {false, true, true}}
	m, _ := mcm.New(ok)
	before := m.Match()
	c := m.Copy()
	for u, v := range before {
		c.Remove(u, v)
	}
	if got := m.Match(); !slices.Equal(got, before) || m.Size() != 3 {
		t.Fatalf("original changed: %v", got)
	}
	cm := c.Match()
	for u, v := range before {
		if cm[u] == v {
			t.Fatalf("copy still pairs %d with %d", u, v)
		}
	}
	cSize := c.Size()
	for u, v := range before {
		m.Remove(u, v)
	}
	m.Match()
	if c.Size() != cSize || !slices.Equal(c.Match(), cm) {
		t.Fatal("copy changed by original")
	}
}

func TestNewErrors(t *testing.T) {
	cases := []struct {
		name string
		f    func() error
		want string
	}{
		{"empty", func() error { _, e := mcm.New([][]bool{}); return e }, "empty"},
		{"empty row", func() error { _, e := mcm.New([][]bool{{}}); return e }, "empty"},
		{"ragged", func() error { _, e := mcm.New([][]bool{{true, true}, {true}}); return e }, "ragged"},
		{"range", func() error {
			_, e := mcm.NewSparse(2, 2, []ap.Pair{{U: 2, V: 0}})
			return e
		}, "out of range"},
		{"duplicate", func() error {
			_, e := mcm.NewSparse(2, 2, []ap.Pair{{U: 0, V: 0}, {U: 0, V: 0}})
			return e
		}, "duplicate"},
	}
	for _, c := range cases {
		err := c.f()
		if err == nil || !strings.HasPrefix(err.Error(), "mcm: ") ||
			!strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: got %v", c.name, err)
		}
	}
}

func TestNewCopiesInput(t *testing.T) {
	adj := [][]bool{{true, true}, {true, true}}
	pairs := []ap.Pair{{U: 0, V: 0}, {U: 0, V: 1}, {U: 1, V: 0}}
	m, _ := mcm.New(adj)
	m.Remove(0, 0)
	m.Remove(1, 1)
	m.Match()
	for _, row := range adj {
		if !row[0] || !row[1] {
			t.Fatalf("adj modified: %v", adj)
		}
	}
	s, _ := mcm.NewSparse(2, 2, pairs)
	s.Remove(0, 0)
	s.Match()
	if pairs[0] != (ap.Pair{U: 0, V: 0}) || pairs[1] != (ap.Pair{U: 0, V: 1}) || pairs[2] != (ap.Pair{U: 1, V: 0}) {
		t.Fatalf("pairs modified: %v", pairs)
	}
}
