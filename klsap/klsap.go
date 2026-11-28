package klsap

import (
	"errors"
	"fmt"

	"github.com/ryanjoneil/ap"
	"github.com/ryanjoneil/ap/lsap"
)

// KLSAP ranks the k best assignments of a linear sum assignment problem.
// Create one with [New] for a complete cost matrix, or with [NewSparse] for a
// sparse edge list, then call Next once for each assignment.
//
// A KLSAP is not safe for concurrent use.
type KLSAP[T ap.Integer] struct {
	root    *lsap.LSAP[T] // unsolved root problem, nil once solved
	k       int           // number of assignments to rank
	r       int           // number of assignments returned so far
	queue   queue[T]      // waiting subproblems
	pending *node[T]      // last node returned, to split on the next call
	seq     uint64        // creation order of the next node, to break ties
	cost    T             // cost of the last assignment returned
	err     error         // sticky error, returned by every later call
}

// New returns a ranking of the k best assignments for the square cost matrix
// costs, where costs[u][v] is the cost of assigning u ∈ U to v ∈ V. It
// validates costs as [lsap.New] does, and returns an error if k < 1. New
// doesn't solve; the first call to Next does.
//
// Every child subproblem forbids an edge, and forbidding an edge in a dense
// lsap raises its big-M, so a ranking has a narrower usable cost range than
// [lsap.New] accepts: roughly, it needs about 3·M of room below the largest
// value of T, where M is lsap's big-M. For int32 and n = 1000, costs up to
// 2,145,338 pass New, but the ranking can fail with [ap.ErrOverflow] after
// its first assignment, while costs up to about 700,000 work.
func New[T ap.Integer](costs [][]T, k int) (*KLSAP[T], error) {
	if k < 1 {
		return nil, errors.New("klsap: k must be at least 1")
	}
	a, err := lsap.New(costs)
	if err != nil {
		return nil, fmt.Errorf("klsap: %w", err)
	}
	return &KLSAP[T]{root: a, k: k}, nil
}

// NewSparse returns a ranking of the k best assignments for the n×n
// assignment problem whose allowed edges are given by edges, where an edge
// {U: u, V: v, Cost: c} allows assigning u ∈ U to v ∈ V at cost c. It
// validates its input as [lsap.NewSparse] does, and returns an error if
// k < 1. NewSparse doesn't solve; the first call to Next does. Subproblems
// that forbid edges can overflow T even when the whole problem does not; see
// [New] and [KLSAP.Next].
func NewSparse[T ap.Integer](n int, edges []ap.Edge[T], k int) (*KLSAP[T], error) {
	if k < 1 {
		return nil, errors.New("klsap: k must be at least 1")
	}
	a, err := lsap.NewSparse(n, edges)
	if err != nil {
		return nil, fmt.Errorf("klsap: %w", err)
	}
	return &KLSAP[T]{root: a, k: k}, nil
}

// Next returns the next assignment in nondecreasing order of cost, and sets
// Cost to its cost. Assignments of equal cost come in an unspecified order,
// which is the same for the same input. No assignment is returned twice. The
// caller owns the returned slice.
//
// Next returns [ap.ErrInfeasible] when there is no further assignment: on
// the first call if the problem has no feasible assignment, after the last
// assignment, and after k assignments. It returns [ap.ErrOverflow] if solving
// a subproblem would overflow T; see [lsap.New] and [lsap.NewSparse]. Once
// Next returns an error, it returns the same error on every later call.
//
// A subproblem that [lsap] reports as ap.ErrOverflow ends the whole ranking
// with ap.ErrOverflow, even if that subproblem was really infeasible, and
// even though earlier calls succeeded.
func (a *KLSAP[T]) Next() (ap.Permutation, error) {
	if a.err != nil {
		return nil, a.err
	}
	if a.r == a.k {
		return nil, a.fail(ap.ErrInfeasible)
	}

	if a.root != nil {
		root := a.root
		a.root = nil
		p, err := root.Assign()
		if err != nil {
			return nil, a.fail(err)
		}
		free := make([]int, len(p))
		for i := range free {
			free[i] = i
		}
		a.offer(&node[T]{a: root, p: p, cost: root.Cost(), free: free})
	} else if a.pending != nil {
		err := a.split(a.pending)
		a.pending = nil
		if err != nil {
			return nil, a.fail(err)
		}
	}

	if a.queue.len() == 0 {
		return nil, a.fail(ap.ErrInfeasible)
	}
	nd := a.queue.popMin()
	a.r++
	a.cost = nd.cost
	if a.r < a.k {
		a.pending = nd
	}
	return append(ap.Permutation(nil), nd.p...), nil
}

// fail makes err sticky and returns it. It releases the waiting subproblems,
// which no later call can use, so that their solvers can be collected.
func (a *KLSAP[T]) fail(err error) error {
	a.err = err
	a.root, a.pending, a.queue = nil, nil, queue[T]{}
	return err
}

// Cost returns the total cost of the assignment last returned by Next. It is
// meaningful only after a successful call to Next.
func (a *KLSAP[T]) Cost() T {
	return a.cost
}
