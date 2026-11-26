package ap

import "errors"

// ErrInfeasible is returned by solvers when no feasible assignment exists.
var ErrInfeasible = errors.New("ap: no feasible assignment")

// Edge is a sparse edge from row U to column V with the given cost.
type Edge[T Integer] struct {
	U, V int
	Cost T
}

// Pair is an assigned (or matched) pair of row U and column V.
type Pair struct {
	U, V int
}
