package lsap

import (
	"errors"
	"fmt"

	"github.com/ryanjoneil/ap"
)

// LSAP is a linear sum assignment problem solver. Create one with [New].
//
// After an initial call to Assign, one can remove edges and call Assign again
// to re-optimize incrementally. Copy duplicates a solver, including its
// current solution, for use in tree search.
//
// An LSAP is not safe for concurrent use. Use Copy to work in parallel.
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
// and computes M with overflow checks.
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

// Assign returns an optimal assignment. It returns [ap.ErrInfeasible] if
// removed edges leave no feasible assignment. A removed edge never appears in
// the returned assignment.
func (a *LSAP[T]) Assign() (ap.Permutation, error) {
	return a.k.assign()
}

// Cost returns the total cost of the assignment found by the last successful
// call to Assign.
func (a *LSAP[T]) Cost() T {
	return a.k.cost()
}

// DualPrices returns the dual prices associated with the assignment
// constraints of the U and V sets. After a successful call to Assign,
// complementary slackness conditions hold.
func (a *LSAP[T]) DualPrices() ap.DualPrices[T] {
	return a.k.dualPrices()
}

// ReducedCost returns the reduced cost of edge (u, v), and false if the edge
// is out of range or removed. After a successful call to Assign, the reduced
// cost of an edge outside the assignment is a lower bound on the additional
// cost of forcing that edge into the assignment.
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
// edge is out of range or already removed.
//
// The initial optimization runs in O(n³) time, after which one can remove
// edges from the resulting assignment and re-solve in O(n²) time per removed
// assigned edge.
func (a *LSAP[T]) Remove(u, v int) {
	a.k.remove(u, v)
}

// Copy returns an independent solver with the same costs, removals and
// current solution. Changes to either solver never affect the other.
func (a *LSAP[T]) Copy() *LSAP[T] {
	return &LSAP[T]{k: a.k.copy()}
}
