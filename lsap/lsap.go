package lsap

import (
	"errors"
	"fmt"

	"github.com/ryanjoneil/ap"
)

// LSAP is a linear sum assignment problem solver. Create one with [New] for a
// complete cost matrix, or with [NewSparse] for a sparse edge list.
//
// One can fix edges, forbidding or requiring them with Fix, before or after
// calling Assign. After a call to Assign, the next one re-optimizes
// incrementally from the current solution. Copy duplicates a solver,
// including its current solution, for use in tree search.
//
// The results reported by Cost, DualPrices and ReducedCost are meaningful
// only after a successful call to Assign. They are undefined before any call
// to Assign, and after a call to Assign that returns an error. After Assign
// returns [ap.ErrOverflow], the solver can't solve again, and neither can
// copies made from it afterwards; see Assign.
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
	fix(u, v int, value bool)
	copy() kernel[T]
}

// New returns a solver for the square cost matrix costs, where costs[u][v] is
// the cost of assigning u ∈ U to v ∈ V. New copies costs, so later changes to
// costs don't affect the solver, and the solver never modifies costs.
//
// New returns an error if costs is empty, ragged or not square, or if its
// costs are too wide for T. Forbidden edges are given the cost
//
//	M = n·(maxC − minC) + maxC + 1
//
// where minC and maxC are the smallest and largest costs. New requires that
// no cost is the smallest value of T, and that M is less than the largest
// value of T. For example, with T = int32 and n = 1000, costs in
// [0, 2145338] are accepted but costs in [0, 2145339] are not, since
// 1001·2145339 + 1 > 2³¹−1.
//
// Passing New doesn't mean that solving fits in T. Assign returns
// [ap.ErrOverflow] if it would overflow T. That happens most often for costs
// above about a third of the largest value of T, or once forbidden edges bring
// M into the costs, but the dual prices can exceed the costs, so it can
// happen for smaller costs too. Use a wider T in that case. To keep a solver
// to fall back on, call Copy before Assign or Fix.
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
// or if a cost is the smallest or largest value of T.
//
// Passing NewSparse doesn't mean that solving fits in T. Assign returns
// [ap.ErrOverflow] if it would overflow T. That happens most often for costs
// above about a fifth of the largest value of T, but the dual prices can
// exceed the costs, so it can happen for smaller costs too. Use a wider T in
// that case. To keep a solver to fall back on, call Copy before Assign or
// Fix.
func NewSparse[T ap.Integer](n int, edges []ap.Edge[T]) (*LSAP[T], error) {
	s, err := newSparse(n, edges)
	if err != nil {
		return nil, err
	}
	return &LSAP[T]{k: s}, nil
}

// Assign returns an optimal assignment that contains every required edge. It
// returns [ap.ErrInfeasible] if the fixes conflict (see [LSAP.Fix]), or if
// forbidden edges, required edges, or edges missing from a sparse solver
// leave no feasible assignment. A forbidden or missing edge never appears in
// the returned assignment.
//
// Assign returns [ap.ErrOverflow] if solving would overflow T; see [New] and
// [NewSparse]. Once it does, every later call to Assign on the solver, or on
// a copy made from it afterwards, returns ap.ErrOverflow too, even after a
// conflicting Fix. If the fixes, or edges missing from a sparse solver, leave
// no feasible assignment, and proving that would overflow T, Assign may
// return ap.ErrOverflow instead of ap.ErrInfeasible.
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
// is out of range, forbidden, missing from a sparse solver, or shares its row
// or column with a required edge other than itself. The reduced cost of a
// required edge is 0. The reduced cost of an edge outside the assignment is a
// lower bound on the additional cost of forcing that edge into the
// assignment. The reduced cost is meaningful only after a successful call to
// Assign, and undefined before any call to Assign or after a call to Assign
// that returns an error. The ok result is always valid.
//
// For example, given some valid primal bound p, an edge (u,v) cannot be part
// of an assignment better than p if:
//
//	if rc, ok := a.ReducedCost(u, v); ok && a.Cost()+rc > p {
//	    a.Fix(u, v, false)
//	}
func (a *LSAP[T]) ReducedCost(u, v int) (T, bool) {
	return a.k.reducedCost(u, v)
}

// Fix forbids edge (u, v) in every later call to Assign if value is false,
// and requires it if value is true. It is a no-op if the edge is out of range
// or already fixed to the same value, and Fix(u, v, false) is a no-op if the
// edge is missing from a sparse solver.
//
// Every other edge in the row or column of a required edge reads as
// forbidden: ReducedCost returns false for it, and it never appears in an
// assignment. Requiring an edge that is already assigned keeps the current
// solution, so the next Assign doesn't re-solve. Otherwise Fix assigns u to
// v, unassigning u's old column and v's old row, and the next Assign
// re-optimizes those, as it does after forbidding an assigned edge. Requiring
// an edge copies no cost rows and runs in O(n) time. It sets the dual price of
// row u, so, as with any other dual price, it can make Assign return
// [ap.ErrOverflow] if that price would leave the range in which solving is
// safe for T.
//
// Fixes that conflict make every later call to Assign return
// [ap.ErrInfeasible]: requiring a forbidden or missing edge, fixing an edge
// both ways in either order, or requiring two edges that share a row or a
// column. There is no way to unfix an edge, so a caller that may need to back
// out should call Copy first.
//
// For a solver created with New, the initial optimization runs in O(n³) time,
// after which one can fix edges of the resulting assignment and re-solve in
// O(n²) time per unassigned row. For a solver created with NewSparse with m
// edges, these are O(n·m·log m) and O(m·log m) time.
func (a *LSAP[T]) Fix(u, v int, value bool) {
	a.k.fix(u, v, value)
}

// Copy returns an independent solver with the same costs, fixes and current
// solution. Changes to either solver never affect the other.
//
// Copy may be called concurrently on the same solver from several
// goroutines. No other method may run concurrently with anything on that
// solver, including Copy.
func (a *LSAP[T]) Copy() *LSAP[T] {
	return &LSAP[T]{k: a.k.copy()}
}
