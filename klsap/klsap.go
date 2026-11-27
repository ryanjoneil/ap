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
// k < 1. NewSparse doesn't solve; the first call to Next does.
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
func (a *KLSAP[T]) Next() (ap.Permutation, error) {
	if a.err != nil {
		return nil, a.err
	}
	if a.r == a.k {
		a.err = ap.ErrInfeasible
		return nil, a.err
	}

	if a.root != nil {
		root := a.root
		a.root = nil
		p, err := root.Assign()
		if err != nil {
			a.err = err
			return nil, a.err
		}
		free := make([]int, len(p))
		for i := range free {
			free[i] = i
		}
		a.offer(&node[T]{a: root, p: p, cost: root.Cost(), free: free})
	} else if a.pending != nil {
		a.split(a.pending)
		a.pending = nil
		if a.err != nil {
			return nil, a.err
		}
	}

	if a.queue.len() == 0 {
		a.err = ap.ErrInfeasible
		return nil, a.err
	}
	nd := a.queue.popMin()
	a.r++
	a.cost = nd.cost
	if a.r < a.k {
		a.pending = nd
	}
	return append(ap.Permutation(nil), nd.p...), nil
}

// Cost returns the total cost of the assignment last returned by Next. It is
// meaningful only after a successful call to Next.
func (a *KLSAP[T]) Cost() T {
	return a.cost
}
