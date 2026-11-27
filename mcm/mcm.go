package mcm

import (
	"fmt"
	"slices"

	"github.com/ryanjoneil/ap"
	"github.com/ryanjoneil/ap/internal/bipartite"
	"github.com/ryanjoneil/ap/internal/graph"
)

// MCM finds a maximum-cardinality matching in a bipartite graph. Create
// one with [New] for an adjacency matrix, or with [NewSparse] for an edge
// list.
//
// One can fix edges, forbidding or requiring them with Fix, before or after
// calling Match or Assign, which re-solve incrementally. Copy duplicates an
// MCM, including its current matching, for use in tree search.
//
// The results reported by Match and Size are meaningful only after a call to
// Match or Assign; Size is 0 before any such call. An MCM is not safe for
// concurrent use. Use Copy to work in parallel.
type MCM struct {
	g      *graph.CSR // owned: Fix reorders Adj within each row
	end    []int      // row u's allowed columns are g.Adj[g.Start[u]:end[u]]
	st     *bipartite.State
	solved bool
	size   int // size of the matching at the last solve

	// fixCol[u] is the column of the required edge in row u, or −1; fixRow[v]
	// is the row of the one in column v, or −1. conflict is set by a Fix that
	// makes the problem infeasible, and is never cleared.
	fixCol, fixRow []int
	conflict       bool
}

// New returns an MCM solver for the bipartite graph in which row u and column v
// are adjacent when adjacency[u][v] is true. The graph may be rectangular.
// New copies adjacency, so later changes to it don't affect the solver.
//
// New returns an error if adjacency is empty or ragged.
func New(adjacency [][]bool) (*MCM, error) {
	g, err := graph.FromAdjacency(adjacency)
	if err != nil {
		return nil, fmt.Errorf("mcm: %w", err)
	}
	return build(g), nil
}

// NewSparse returns an MCM solver for the bipartite graph with nU rows, nV
// columns and the given edges, where a pair {U: u, V: v} joins row u to
// column v. NewSparse copies edges, so later changes to edges don't affect
// the solver. The order of edges does not matter.
//
// NewSparse returns an error if nU or nV is less than 1, or if an edge is out
// of range or duplicated.
func NewSparse(nU, nV int, edges []ap.Pair) (*MCM, error) {
	g, err := graph.FromPairs(nU, nV, edges)
	if err != nil {
		return nil, fmt.Errorf("mcm: %w", err)
	}
	return build(g), nil
}

func build(g *graph.CSR) *MCM {
	m := &MCM{
		g:      g,
		end:    slices.Clone(g.Start[1:]),
		st:     bipartite.New(g.NU, g.NV),
		fixCol: minusOnes(g.NU),
		fixRow: minusOnes(g.NV),
	}
	m.st.Greedy(g.Start[:g.NU], m.end, g.Adj)
	return m
}

func minusOnes(n int) []int {
	s := make([]int, n)
	for i := range s {
		s[i] = -1
	}
	return s
}

func (m *MCM) solve() {
	if !m.solved {
		m.size = m.st.Maximize(m.g.Start[:m.g.NU], m.end, m.g.Adj)
		m.solved = true
	}
}

// Match returns a maximum matching as a slice of length |U|: element u is the
// column matched to row u, or −1 if u is unmatched. It solves first if edges
// were fixed since the last solve, and returns a new slice each
// call. The matching contains every required edge. Match returns
// (nil, [ap.ErrInfeasible]) if the fixes conflict (see [MCM.Fix]).
func (m *MCM) Match() ([]int, error) {
	if m.conflict {
		return nil, ap.ErrInfeasible
	}
	m.solve()
	return slices.Clone(m.st.MateU), nil
}

// Size returns the size of the matching found by the last call to Match or
// Assign. It is valid only after one of those calls, and is 0 before.
func (m *MCM) Size() int { return m.size }

// Assign returns the perfect matching as a permutation. It returns
// [ap.ErrInfeasible] if the fixes conflict, or unless |U| = |V| and every row
// can be matched. Assign always solves, so Size is valid afterwards even when
// it returns an error.
func (m *MCM) Assign() (ap.Permutation, error) {
	if m.conflict {
		return nil, ap.ErrInfeasible
	}
	m.solve()
	if m.g.NU != m.g.NV || m.size != m.g.NU {
		return nil, ap.ErrInfeasible
	}
	return ap.Permutation(slices.Clone(m.st.MateU)), nil
}

// Fix forbids the edge from row u to column v in every later solve if value
// is false, and requires it in the matching if value is true. It is a no-op
// if u or v is out of range or the edge is already fixed to the same value,
// and Fix(u, v, false) is a no-op if the edge is missing.
//
// Every other edge in the row or column of a required edge reads as forbidden.
// Forbidding an edge that is not in the current matching leaves the maximum
// matching unchanged, so the next Match does not need to re-solve. Requiring
// an edge forbids its row and column neighbors, and re-solves if that
// unmatches anything. Forbidding an edge takes time linear in the degree of
// row u. Requiring one also scans the allowed edges of every row, to forbid
// those in column v.
//
// Fixes that conflict make every later call to Match and Assign return
// [ap.ErrInfeasible], and Size return 0: requiring a forbidden or missing edge,
// fixing an edge both ways in either order, or requiring two edges that share
// a row or a column. There is no way to unfix an edge, so a caller that may
// need to back out should call Copy first.
func (m *MCM) Fix(u, v int, value bool) {
	if u < 0 || u >= m.g.NU || v < 0 || v >= m.g.NV || m.conflict {
		return
	}
	if !value {
		if m.fixCol[u] == v {
			m.setConflict()
			return
		}
		m.forbid(u, v)
		return
	}
	if m.fixCol[u] == v {
		return
	}
	if m.fixCol[u] >= 0 || m.fixRow[v] >= 0 ||
		!slices.Contains(m.g.Adj[m.g.Start[u]:m.end[u]], v) {
		m.setConflict()
		return
	}
	m.fixCol[u], m.fixRow[v] = v, u
	// Keep only column v in row u's window, in one pass.
	lo := m.g.Start[u]
	e := lo + slices.Index(m.g.Adj[lo:m.end[u]], v)
	m.g.Adj[lo], m.g.Adj[e] = m.g.Adj[e], m.g.Adj[lo]
	m.end[u] = lo + 1
	if mate := m.st.MateU[u]; mate >= 0 && mate != v {
		m.st.Unmatch(u)
		m.solved = false
	}
	for w := range m.g.NU {
		if w != u {
			m.forbid(w, v)
		}
	}
}

func (m *MCM) setConflict() {
	m.conflict = true
	m.size = 0
}

// forbid drops the edge from row u to column v from the allowed edges. It does
// nothing if the edge is missing or already forbidden.
func (m *MCM) forbid(u, v int) {
	lo, hi := m.g.Start[u], m.end[u]
	adj := m.g.Adj
	for e := lo; e < hi; e++ {
		if adj[e] != v {
			continue
		}
		// Swap with the last allowed entry and shrink the window; the
		// forbidden edge stays in Adj past end[u].
		adj[e], adj[hi-1] = adj[hi-1], adj[e]
		m.end[u]--
		if m.st.MateU[u] == v {
			m.st.Unmatch(u)
			m.solved = false
		}
		return
	}
}

// Copy returns an independent solver with the same graph and current
// matching. Changes to either solver never affect the other.
func (m *MCM) Copy() *MCM {
	g := *m.g
	g.Start = slices.Clone(m.g.Start)
	g.Adj = slices.Clone(m.g.Adj)
	c := *m
	c.g = &g
	c.end = slices.Clone(m.end)
	c.fixCol = slices.Clone(m.fixCol)
	c.fixRow = slices.Clone(m.fixRow)
	c.st = m.st.Clone()
	return &c
}
