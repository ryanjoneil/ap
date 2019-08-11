package ap

// An Int64Coster provides an int64 cost value. This is usually the objective of
// value of a particular assignment.
type Int64Coster interface {
	Cost() int64
}

// An Int64ReducedCoster provides a method for computing the reduced cost of
// assigning u ∈ U to v ∈ V, where u and v are both integers from 0 to n-1. The
// reduced cost of a basic edge (already part of an assignment) is zero, since
// it does not change the solution. Introducing a nonbasic edge (not in the
// assignment) may change the resulting assignment's overall cost.
type Int64ReducedCoster interface {
	ReducedCost(u, v int) int64
}
