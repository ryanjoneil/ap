package ap

// An Assignment is a matching of elements from one set U = {0,...,n-1} to
// another set V = {0,...,n-1}. It is represented as a permutation P, where
// P[u] = v means that u ∈ U is assigned to v ∈ V.
//
//     a := ap.Assignment{1, 0, 2} // Assign 0 to 1, 1 to 0, and 2 to itself.
type Assignment []int

// An Assigner creates a mapping between two sets U = V = {0,...,n-1}.
type Assigner interface {
	Assign() Assignment
}

// An Int64Coster provides an int64 cost value. This is usually the objective of
// value of a particular assignment.
type Int64Coster interface {
	Cost() int64
}

// Int64DualPrices are dual prices for the assignment constraints corresponding
// to the U and V sets, respectively.
type Int64DualPrices struct {
	U []int64
	V []int64
}

// An Int64DualPricer provides dual prices on the assignment constraints
// associated with sets U and V. A dual price is the value of a unit of slack on
// a binding constraint.
type Int64DualPricer interface {
	DualPrices() Int64DualPrices
}

// An Int64ReducedCoster provides a method for computing the reduced cost of
// assigning u ∈ U to v ∈ V, where u and v are both integers from 0 to n-1. The
// reduced cost of a basic edge (already part of an assignment) is zero, since
// it does not change the solution. Introducing a nonbasic edge (not in the
// assignment) may change the resulting assignment's overall cost.
type Int64ReducedCoster interface {
	ReducedCost(u, v int) int64
}
