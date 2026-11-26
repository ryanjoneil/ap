package lsap

import (
	"errors"
	"fmt"

	"github.com/ryanjoneil/ap"
)

// LSAP is a linear sum assignment problem solver. Create one with [New] for a
// complete cost matrix, or with [NewSparse] for a sparse edge list.
//
// After an initial call to Assign, one can remove edges and call Assign again
// to re-optimize incrementally. Copy duplicates a solver, including its
// current solution, for use in tree search.
//
// The results reported by Cost, DualPrices and ReducedCost are meaningful
// only after a successful call to Assign. They are undefined before any call
// to Assign, and after a call to Assign that returns an error.
//
// An LSAP is not safe for concurrent use, with one exception: several
// goroutines may call Copy on the same solver at once, so that a solved parent
// can seed parallel work. No other method may run concurrently with any
// method, including Copy, on the same solver.
type LSAP[T ap.Integer] struct{ k kernel[T] }

// kernel is a solver implementation behind an LSAP.
type kernel[T ap.Integer] interface {
	assign() (ap.Permutation, error)
	cost() T
	dualPrices() ap.DualPrices[T]
	reducedCost(u, v int) (T, bool)
	remove(u, v int)
	copy() kernel[T]
}

// New returns a solver for the square cost matrix costs, where costs[u][v] is
// the cost of assigning u ∈ U to v ∈ V. New copies costs, so later changes to
// costs don't affect the solver, and the solver never modifies costs.
//
// New returns an error if costs is empty, ragged or not square, or if its
// costs could overflow T while solving. Removed edges are given the cost
//
//	M = n·(maxC − minC) + maxC + 1
//
// where minC and maxC are the smallest and largest costs. New requires
//
//	4·(n+1)·max(|minC|, |M|) ≤ the largest value of T
//
// and computes M with overflow checks. The bound is restrictive for narrow
// types: for example, with T = int32 and n = 1000, costs in [0, 535] are
// accepted but costs in [0, 536] are not, since
// 4·1001·(1000·536 + 536 + 1) > 2³¹−1. Use int64, or NewSparse (which needs
// no M), for wider cost ranges.
func New[T ap.Integer](costs [][]T) (*LSAP[T], error) {
	n := len(costs)
	if n == 0 {
		return nil, errors.New("lsap: empty cost matrix")
	}
	for u, row := range costs {
		if len(row) != n {
			return nil, fmt.Errorf(
				"lsap: cost matrix not square: row %d has %d costs; want %d",
				u, len(row), n,
			)
		}
	}

	d, err := newDense(costs)
	if err != nil {
		return nil, err
	}
	return &LSAP[T]{k: d}, nil
}

// NewSparse returns a solver for the n×n assignment problem whose allowed
// edges are given by edges, where an edge {U: u, V: v, Cost: c} allows
// assigning u ∈ U to v ∈ V at cost c. Pairs not in edges are forbidden, so
// Assign returns [ap.ErrInfeasible] if the edges admit no perfect assignment.
// NewSparse copies edges, so later changes to edges don't affect the solver,
// and the solver never modifies edges. The order of edges does not matter.
//
// NewSparse returns an error if n < 1, if n > math.MaxInt32 (the search
// stores column indices as int32), if an edge is out of range or duplicated,
// or if its costs could overflow T while solving. It requires
//
//	4·(n+1)·max |c| ≤ the largest value of T
//
// over the costs c of all edges.
func NewSparse[T ap.Integer](n int, edges []ap.Edge[T]) (*LSAP[T], error) {
	s, err := newSparse(n, edges)
	if err != nil {
		return nil, err
	}
	return &LSAP[T]{k: s}, nil
}

// Assign returns an optimal assignment. It returns [ap.ErrInfeasible] if
// removed edges, or edges missing from a sparse solver, leave no feasible
// assignment. A removed or missing edge never appears in the returned
// assignment.
func (a *LSAP[T]) Assign() (ap.Permutation, error) {
	return a.k.assign()
}

// Cost returns the total cost of the assignment returned by Assign. It is
// meaningful only after a successful call to Assign, and undefined before any
// call to Assign or after a call to Assign that returns an error.
func (a *LSAP[T]) Cost() T {
	return a.k.cost()
}

// DualPrices returns the dual prices associated with the assignment
// constraints of the U and V sets, which satisfy complementary slackness with
// the assignment returned by Assign. They are meaningful only after a
// successful call to Assign, and undefined before any call to Assign or after
// a call to Assign that returns an error.
func (a *LSAP[T]) DualPrices() ap.DualPrices[T] {
	return a.k.dualPrices()
}

// ReducedCost returns the reduced cost of edge (u, v), and false if the edge
// is out of range, removed, or missing from a sparse solver. The reduced cost of an edge outside the
// assignment is a lower bound on the additional cost of forcing that edge into
// the assignment. The reduced cost is meaningful only after a successful call
// to Assign, and undefined before any call to Assign or after a call to Assign
// that returns an error. The ok result is always valid.
//
// For example, given some valid primal bound p, an edge (u,v) cannot be part
// of an assignment better than p if:
//
//	if rc, ok := a.ReducedCost(u, v); ok && a.Cost()+rc > p {
//	    a.Remove(u, v)
//	}
func (a *LSAP[T]) ReducedCost(u, v int) (T, bool) {
	return a.k.reducedCost(u, v)
}

// Remove forbids edge (u, v) in later calls to Assign. It is a no-op if the
// edge is out of range, missing or already removed.
//
// For a solver created with New, the initial optimization runs in O(n³) time,
// after which one can remove edges from the resulting assignment and re-solve
// in O(n²) time per removed assigned edge. For a solver created with
// NewSparse with m edges, these are O(n·m·log m) and O(m·log m) time.
func (a *LSAP[T]) Remove(u, v int) {
	a.k.remove(u, v)
}

// Copy returns an independent solver with the same costs, removals and
// current solution. Changes to either solver never affect the other.
//
// Copy may be called concurrently on the same solver from several
// goroutines. No other method may run concurrently with anything on that
// solver, including Copy.
func (a *LSAP[T]) Copy() *LSAP[T] {
	return &LSAP[T]{k: a.k.copy()}
}
