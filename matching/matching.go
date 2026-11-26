package matching

import (
	"fmt"
	"slices"

	"github.com/ryanjoneil/ap"
	"github.com/ryanjoneil/ap/internal/bipartite"
	"github.com/ryanjoneil/ap/internal/graph"
)

// Matcher finds a maximum-cardinality matching in a bipartite graph. Create
// one with [New] for an adjacency matrix, or with [NewSparse] for an edge
// list.
//
// After a call to Match or Assign, one can remove edges and call either again
// to re-solve incrementally. Copy duplicates a matcher, including its current
// matching, for use in tree search.
//
// The results reported by Match and Size are meaningful only after a call to
// Match or Assign; Size is 0 before any such call. A Matcher is not safe for
// concurrent use. Use Copy to work in parallel.
type Matcher struct {
	g      *graph.CSR // owned: Remove reorders Adj within each row
	end    []int      // row u's allowed columns are g.Adj[g.Start[u]:end[u]]
	st     *bipartite.State
	solved bool
	size   int // size of the matching at the last solve
}

// New returns a matcher for the bipartite graph in which row u and column v
// are adjacent when adjacency[u][v] is true. The graph may be rectangular.
// New copies adjacency, so later changes to it don't affect the matcher.
//
// New returns an error if adjacency is empty or ragged.
func New(adjacency [][]bool) (*Matcher, error) {
	g, err := graph.FromAdjacency(adjacency)
	if err != nil {
		return nil, fmt.Errorf("matching: %w", err)
	}
	return build(g), nil
}

// NewSparse returns a matcher for the bipartite graph with nU rows, nV
// columns and the given edges, where a pair {U: u, V: v} joins row u to
// column v. NewSparse copies edges, so later changes to edges don't affect
// the matcher. The order of edges does not matter.
//
// NewSparse returns an error if nU or nV is less than 1, or if an edge is out
// of range or duplicated.
func NewSparse(nU, nV int, edges []ap.Pair) (*Matcher, error) {
	g, err := graph.FromPairs(nU, nV, edges)
	if err != nil {
		return nil, fmt.Errorf("matching: %w", err)
	}
	return build(g), nil
}

func build(g *graph.CSR) *Matcher {
	m := &Matcher{g: g, end: slices.Clone(g.Start[1:]), st: bipartite.New(g.NU, g.NV)}
	m.st.Greedy(g.Start[:g.NU], m.end, g.Adj)
	return m
}

func (m *Matcher) solve() {
	if !m.solved {
		m.size = m.st.Maximize(m.g.Start[:m.g.NU], m.end, m.g.Adj)
		m.solved = true
	}
}

// Match returns a maximum matching as a slice of length |U|: element u is the
// column matched to row u, or −1 if u is unmatched. It solves first if edges
// were removed since the last solve. Each call returns a new slice.
func (m *Matcher) Match() []int {
	m.solve()
	return slices.Clone(m.st.MateU)
}

// Size returns the size of the matching found by the last call to Match or
// Assign. It is valid only after one of those calls, and is 0 before.
func (m *Matcher) Size() int { return m.size }

// Assign returns the perfect matching as a permutation. It returns
// [ap.ErrInfeasible] unless |U| = |V| and every row can be matched. Assign
// always solves, so Size is valid afterwards even when it returns an error.
func (m *Matcher) Assign() (ap.Permutation, error) {
	m.solve()
	if m.g.NU != m.g.NV || m.size != m.g.NU {
		return nil, ap.ErrInfeasible
	}
	return ap.Permutation(slices.Clone(m.st.MateU)), nil
}

// Remove deletes the edge from row u to column v. It does nothing if u or v
// is out of range or the edge is missing or already removed. Removing an
// edge that is not in the current matching leaves the maximum matching
// unchanged, so the next Match does not need to re-solve.
func (m *Matcher) Remove(u, v int) {
	if u < 0 || u >= m.g.NU || v < 0 || v >= m.g.NV {
		return
	}
	lo, hi := m.g.Start[u], m.end[u]
	adj := m.g.Adj
	for e := lo; e < hi; e++ {
		if adj[e] != v {
			continue
		}
		// Swap with the last allowed entry and shrink the window; the
		// removed edge stays in Adj past end[u].
		adj[e], adj[hi-1] = adj[hi-1], adj[e]
		m.end[u]--
		if m.st.MateU[u] == v {
			m.st.Unmatch(u)
			m.solved = false
		}
		return
	}
}

// Copy returns an independent matcher with the same graph and current
// matching. Changes to either matcher never affect the other.
func (m *Matcher) Copy() *Matcher {
	g := *m.g
	g.Start = slices.Clone(m.g.Start)
	g.Adj = slices.Clone(m.g.Adj)
	c := *m
	c.g = &g
	c.end = slices.Clone(m.end)
	c.st = m.st.Clone()
	return &c
}
