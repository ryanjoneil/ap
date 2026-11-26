package bipartite

import (
	"math/rand"
	"slices"
	"testing"

	"github.com/ryanjoneil/ap/internal/bruteforce"
	"github.com/ryanjoneil/ap/internal/graph"
)

func randOK(r *rand.Rand, nU, nV int, p float64) [][]bool {
	ok := make([][]bool, nU)
	for u := range ok {
		ok[u] = make([]bool, nV)
		for v := range ok[u] {
			ok[u][v] = r.Float64() < p
		}
	}
	return ok
}

func csr(t *testing.T, ok [][]bool) (start, end, adj []int) {
	t.Helper()
	g, err := graph.FromAdjacency(ok)
	if err != nil {
		t.Fatal(err)
	}
	return g.Start[:g.NU], g.Start[1:], g.Adj
}

func checkValid(t *testing.T, s *State, ok [][]bool) {
	t.Helper()
	n := 0
	for u, v := range s.MateU {
		if v < 0 {
			continue
		}
		n++
		if !ok[u][v] || s.MateV[v] != u {
			t.Fatalf("invalid pair u=%d v=%d", u, v)
		}
	}
	for v, u := range s.MateV {
		if u >= 0 && s.MateU[u] != v {
			t.Fatalf("inconsistent MateV[%d]=%d", v, u)
		}
	}
	if s.Size() != n {
		t.Fatalf("Size %d want %d", s.Size(), n)
	}
}

func TestMaximizeBruteForce(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	ps := []float64{0.2, 0.5, 0.8}
	for i := 0; i < 2000; i++ {
		nU, nV := 1+r.Intn(7), 1+r.Intn(7)
		ok := randOK(r, nU, nV, ps[i%3])
		start, end, adj := csr(t, ok)
		s := New(nU, nV)
		got := s.Maximize(start, end, adj)
		if want := bruteforce.MaxMatching(ok); got != want {
			t.Fatalf("iter %d: got %d want %d", i, got, want)
		}
		checkValid(t, s, ok)
	}
}

func kuhn(ok [][]bool) int {
	nV := len(ok[0])
	mate := make([]int, nV)
	for i := range mate {
		mate[i] = -1
	}
	var try func(u int, seen []bool) bool
	try = func(u int, seen []bool) bool {
		for v := 0; v < nV; v++ {
			if ok[u][v] && !seen[v] {
				seen[v] = true
				if mate[v] < 0 || try(mate[v], seen) {
					mate[v] = u
					return true
				}
			}
		}
		return false
	}
	n := 0
	for u := range ok {
		if try(u, make([]bool, nV)) {
			n++
		}
	}
	return n
}

func TestMaximizeKuhn(t *testing.T) {
	r := rand.New(rand.NewSource(2))
	ps := []float64{0.05, 0.1, 0.3, 0.6}
	for i := 0; i < 1000; i++ {
		nU, nV := 1+r.Intn(40), 1+r.Intn(40)
		ok := randOK(r, nU, nV, ps[i%4])
		start, end, adj := csr(t, ok)
		s := New(nU, nV)
		if i%2 == 0 {
			s.Greedy(start, end, adj)
		}
		got := s.Maximize(start, end, adj)
		if want := kuhn(ok); got != want {
			t.Fatalf("iter %d: got %d want %d", i, got, want)
		}
		checkValid(t, s, ok)
	}
}

func TestWarmStart(t *testing.T) {
	r := rand.New(rand.NewSource(3))
	for i := 0; i < 500; i++ {
		nU, nV := 1+r.Intn(7), 1+r.Intn(7)
		full := randOK(r, nU, nV, 0.5)
		sub := make([][]bool, nU)
		for u := range sub {
			sub[u] = make([]bool, nV)
			for v := range sub[u] {
				sub[u][v] = full[u][v] && r.Intn(2) == 0
			}
		}
		s := New(nU, nV)
		a, b, c := csr(t, sub)
		s.Greedy(a, b, c)
		checkValid(t, s, sub)
		before := slices.Clone(s.MateU)
		start, end, adj := csr(t, full)
		got := s.Maximize(start, end, adj)
		if want := bruteforce.MaxMatching(full); got != want {
			t.Fatalf("iter %d: got %d want %d (start %v)", i, got, want, before)
		}
		checkValid(t, s, full)
	}
}

func TestGreedyKeepsExisting(t *testing.T) {
	ok := [][]bool{{true, true}, {true, true}, {true, false}}
	start, end, adj := csr(t, ok)
	s := New(3, 2)
	s.MateU[0], s.MateV[1] = 1, 0
	s.Greedy(start, end, adj)
	if !slices.Equal(s.MateU, []int{1, 0, -1}) {
		t.Errorf("MateU %v", s.MateU)
	}
}

func TestUnmatchThenMaximize(t *testing.T) {
	r := rand.New(rand.NewSource(4))
	for i := 0; i < 500; i++ {
		nU, nV := 1+r.Intn(7), 1+r.Intn(7)
		ok := randOK(r, nU, nV, 0.6)
		start, end, adj := csr(t, ok)
		s := New(nU, nV)
		s.Maximize(start, end, adj)
		u := r.Intn(nU)
		old := s.MateU[u]
		s.Unmatch(u)
		s.Unmatch(u) // no-op when free
		if old < 0 {
			continue
		}
		ok[u][old] = false
		start, end, adj = csr(t, ok)
		got := s.Maximize(start, end, adj)
		if want := bruteforce.MaxMatching(ok); got != want {
			t.Fatalf("iter %d: got %d want %d", i, got, want)
		}
		checkValid(t, s, ok)
	}
}

func TestWindows(t *testing.T) {
	// Shrunk windows: end[u] cuts off trailing edges of the shared adj.
	ok := [][]bool{{true, true, true}, {true, true, true}}
	start, _, adj := csr(t, ok)
	s := New(2, 3)
	if got := s.Maximize(start, []int{1, 1}, adj); got != 1 {
		t.Errorf("got %d want 1", got)
	}
}

func TestClone(t *testing.T) {
	ok := [][]bool{{true, true}, {true, true}}
	start, end, adj := csr(t, ok)
	s := New(2, 2)
	s.Maximize(start, end, adj)
	orig := slices.Clone(s.MateU)
	c := s.Clone()
	c.Unmatch(0)
	c.Unmatch(1)
	if !slices.Equal(s.MateU, orig) || c.Size() != 0 || s.Size() != 2 {
		t.Errorf("clone not independent: %v %v", s.MateU, c.MateU)
	}
	if got := c.Maximize(start, end, adj); got != 2 {
		t.Errorf("clone maximize %d", got)
	}
}

func TestLargeNoOverflow(t *testing.T) {
	// A long path forces augmenting paths of length ~n; recursion would be deep.
	const n = 100000
	start := make([]int, n+1)
	adj := make([]int, 0, 2*n)
	for u := 0; u < n; u++ {
		start[u] = len(adj)
		adj = append(adj, u)
		if u+1 < n {
			adj = append(adj, u+1)
		}
	}
	start[n] = len(adj)
	s := New(n, n)
	// Bad warm start: row u takes column u+1, leaving row n-1 and column 0 free.
	for u := 0; u+1 < n; u++ {
		s.MateU[u], s.MateV[u+1] = u+1, u
	}
	if got := s.Maximize(start[:n], start[1:], adj); got != n {
		t.Errorf("got %d want %d", got, n)
	}
}
