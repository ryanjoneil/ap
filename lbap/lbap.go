package lbap

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/ryanjoneil/ap"
	"github.com/ryanjoneil/ap/internal/bipartite"
	"github.com/ryanjoneil/ap/internal/graph"
)

// LBAP is a linear bottleneck assignment problem solver. Create one with [New]
// for a complete cost matrix, or with [NewSparse] for a sparse edge list.
//
// After an initial call to Assign, one can remove edges and call Assign again
// to re-optimize incrementally. Copy duplicates a solver, including its
// current solution, for use in tree search.
//
// The result reported by Cost is meaningful only after a successful call to
// Assign. It is undefined before any call to Assign, and after a call to
// Assign that returns an error.
//
// An LBAP is not safe for concurrent use. Use Copy to work in parallel.
type LBAP[T ap.Integer] struct {
	n      int
	levels []T   // sorted distinct costs of the original edges; shared by copies
	start  []int // row u's entries are adj[start[u]:start[u+1]]; shared by copies

	// Row u's live entries are adj[start[u]:start[u]+rowLen[u]], sorted by
	// (cost, column), with cost parallel to adj. Removed entries follow them,
	// still in cost order.
	adj    []int
	cost   []T
	rowLen []int

	// end[u] − start[u] is the number of live entries of row u with cost at
	// most the threshold of the last test, so the allowed columns for that
	// test are adj[start[u]:end[u]].
	end []int
	st  *bipartite.State

	valid bool // st.MateU is an optimal assignment for the current edges
	level int  // index in levels of the last optimum; a lower bound on the next
}

// New returns a solver for the square cost matrix costs, where costs[u][v] is
// the cost of assigning u ∈ U to v ∈ V. New copies costs, so later changes to
// costs don't affect the solver, and the solver never modifies costs.
//
// New returns an error if costs is empty, ragged or not square. Costs are
// only compared, never added, so any values of T are allowed.
func New[T ap.Integer](costs [][]T) (*LBAP[T], error) {
	g, c, err := graph.FromMatrix(costs)
	if err != nil {
		return nil, fmt.Errorf("lbap: %w", err)
	}
	return build(g, c), nil
}

// NewSparse returns a solver for the n×n bottleneck assignment problem whose
// allowed edges are given by edges, where an edge {U: u, V: v, Cost: c} allows
// assigning u ∈ U to v ∈ V at cost c. Pairs not in edges are forbidden, so
// Assign returns [ap.ErrInfeasible] if the edges admit no perfect assignment.
// NewSparse copies edges, so later changes to edges don't affect the solver,
// and the solver never modifies edges. The order of edges does not matter.
//
// NewSparse returns an error if n < 1, or if an edge is out of range or
// duplicated. Costs are only compared, so any values of T are allowed.
func NewSparse[T ap.Integer](n int, edges []ap.Edge[T]) (*LBAP[T], error) {
	g, c, err := graph.FromEdges(n, edges)
	if err != nil {
		return nil, fmt.Errorf("lbap: %w", err)
	}
	return build(g, c), nil
}

// build takes ownership of g.Adj and costs, which the graph builders allocate
// fresh, and re-sorts each row by (cost, column) in place.
func build[T ap.Integer](g *graph.CSR, costs []T) *LBAP[T] {
	n := g.NU
	type entry struct {
		c T
		v int
	}
	var buf []entry
	rowLen := make([]int, n)
	for u := range n {
		lo, hi := g.Start[u], g.Start[u+1]
		buf = buf[:0]
		for e := lo; e < hi; e++ {
			buf = append(buf, entry{costs[e], g.Adj[e]})
		}
		// Columns are distinct within a row, so this order is total and the
		// result does not depend on the sort's stability.
		slices.SortFunc(buf, func(a, b entry) int {
			if c := cmp.Compare(a.c, b.c); c != 0 {
				return c
			}
			return cmp.Compare(a.v, b.v)
		})
		for k, x := range buf {
			costs[lo+k], g.Adj[lo+k] = x.c, x.v
		}
		rowLen[u] = hi - lo
	}

	levels := slices.Clone(costs)
	slices.Sort(levels)
	levels = slices.Clip(slices.Compact(levels))

	return &LBAP[T]{
		n:      n,
		levels: levels,
		start:  g.Start,
		adj:    g.Adj,
		cost:   costs,
		rowLen: rowLen,
		end:    slices.Clone(g.Start[:n]), // empty windows
		st:     bipartite.New(n, n),
	}
}

// test reports whether the edges with cost at most levels[L] contain a
// perfect matching, and leaves a maximum matching on those edges in a.st.
//
// It warm-starts from the matching left by the previous test. When a row's
// window shrinks, the row is unmatched if its mate is among the entries
// dropped from the window, adj[end'[u]:end[u]], which are exactly its live
// edges with cost above levels[L]. Scanning only the dropped entries finds
// the mate's cost without storing it per row.
func (a *LBAP[T]) test(L int) bool {
	thr := a.levels[L]
	for u := range a.n {
		lo := a.start[u]
		// Binary search for the first live entry with cost > thr.
		i, j := lo, lo+a.rowLen[u]
		for i < j {
			h := int(uint(i+j) >> 1)
			if a.cost[h] <= thr {
				i = h + 1
			} else {
				j = h
			}
		}
		if i < a.end[u] {
			if m := a.st.MateU[u]; m >= 0 && slices.Contains(a.adj[i:a.end[u]], m) {
				a.st.Unmatch(u)
			}
		}
		a.end[u] = i
	}
	return a.st.Maximize(a.start[:a.n], a.end, a.adj) == a.n
}

// Assign returns an assignment whose largest cost is as small as possible. It
// returns [ap.ErrInfeasible] if removed edges, or edges missing from a sparse
// solver, leave no feasible assignment. A removed or missing edge never
// appears in the returned assignment. Among optimal assignments, which one is
// returned is deterministic but otherwise unspecified.
//
// Assign is the threshold algorithm: it binary searches the distinct costs
// for the smallest threshold whose edges contain a perfect matching. After
// removals it searches only from the previous optimum upward, since removing
// edges never decreases the bottleneck.
func (a *LBAP[T]) Assign() (ap.Permutation, error) {
	if !a.valid {
		top := len(a.levels) - 1
		if top < 0 || !a.test(top) {
			return nil, ap.ErrInfeasible
		}
		// Invariant: levels[hi] is feasible; levels below lo are not.
		//
		// Departure from Algorithm 6.1, which tests the median of the cost
		// entries strictly between the current bounds: this bisects the
		// index range of the distinct cost levels instead, so each test
		// halves the number of candidate values whatever their
		// multiplicities. It also tests the largest level explicitly rather
		// than assuming it feasible, because a sparse or reduced graph may
		// have no perfect matching at all.
		lo, hi := a.level, top
		for lo < hi {
			mid := lo + (hi-lo)/2
			if a.test(mid) {
				hi = mid
			} else {
				lo = mid + 1
			}
		}
		// The last test may have been an infeasible one below hi; re-test
		// at hi so that a.st holds a perfect matching on its edges. The
		// windows only grow, so the warm start is kept.
		a.test(hi)
		a.level = hi
		a.valid = true
	}
	return ap.Permutation(slices.Clone(a.st.MateU)), nil
}

// Cost returns the largest cost in the assignment returned by Assign: the
// bottleneck value. It is meaningful only after a successful call to Assign,
// and undefined before any call to Assign or after a call to Assign that
// returns an error.
func (a *LBAP[T]) Cost() T {
	if len(a.levels) == 0 {
		var zero T
		return zero
	}
	return a.levels[a.level]
}

// Remove forbids edge (u, v) in later calls to Assign. It is a no-op if the
// edge is out of range, missing or already removed.
//
// Removing an edge outside the current assignment keeps that assignment,
// which stays optimal, so the next Assign returns at once. Removing an
// assigned edge makes the next Assign re-solve, searching only thresholds at
// or above the previous bottleneck and warm-starting from the rest of the
// assignment. Remove takes time linear in the degree of row u.
func (a *LBAP[T]) Remove(u, v int) {
	if u < 0 || u >= a.n || v < 0 || v >= a.n {
		return
	}
	lo := a.start[u]
	last := lo + a.rowLen[u] - 1
	e := slices.Index(a.adj[lo:last+1], v)
	if e < 0 {
		return
	}
	e += lo
	// Shift the entry past the live prefix; both parts stay in cost order.
	c := a.cost[e]
	copy(a.adj[e:last], a.adj[e+1:last+1])
	copy(a.cost[e:last], a.cost[e+1:last+1])
	a.adj[last], a.cost[last] = v, c
	a.rowLen[u]--
	if e < a.end[u] {
		a.end[u]--
	}
	if a.st.MateU[u] == v {
		a.st.Unmatch(u)
		a.valid = false
	}
}

// Copy returns an independent solver with the same costs, removals and
// current solution. Changes to either solver never affect the other.
func (a *LBAP[T]) Copy() *LBAP[T] {
	c := *a // levels and start are never modified, so they are shared
	c.adj = slices.Clone(a.adj)
	c.cost = slices.Clone(a.cost)
	c.rowLen = slices.Clone(a.rowLen)
	c.end = slices.Clone(a.end)
	c.st = a.st.Clone()
	return &c
}
