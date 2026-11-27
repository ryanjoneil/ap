package ap

import "errors"

// ErrInfeasible is returned by solvers when no feasible assignment exists.
var ErrInfeasible = errors.New("ap: no feasible assignment")

// ErrOverflow is returned by solvers when solving would overflow the cost type
// T. Use a wider T.
var ErrOverflow = errors.New("ap: arithmetic overflow")

// Edge is a sparse edge from row U to column V with the given cost.
type Edge[T Integer] struct {
	U, V int
	Cost T
}

// Pair is an unweighted edge: it allows assigning (or matching) row U to
// column V, without a cost.
type Pair struct {
	U, V int
}
