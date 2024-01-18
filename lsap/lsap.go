package lsap

import "github.com/ryanjoneil/ap"

// LSAP implements the Assigner interface.
type LSAP[T ap.Integer] struct {
	m T   // A large cost to avoid using edges
	n int // n of assignment problem

	a [][]T // a[i][j] = cost of assigning row i to column j
	u []T   // u[i] = dual price for row i
	v []T   // v[j] = dual price for column j
	z T     // Objective value

	f    []int // f[i] = column assigned to row i, -1 if unassigned
	fBar []int // fBar[j] = row assigned to column j, -1 if unassigned
	p    []int // First column of row i not yet examined in phase 2
	c    []int // c[j] = row preceding column j in current alternating path
	pi   []T   // pi[j] = min { a[i,j] - u[i] - v[j] | i in lr, i != fBar[j] }
}

// Assign returns an optimal assignment for an LSAP.
func (a *LSAP[T]) Assign() ap.Permutation {
	a.solve()
	p := make(ap.Permutation, a.n)
	copy(p, a.f)
	return p
}

// Cost returns the objective value of an LSAP. If Assign is called prior to
// calling Cost, then the value is minimal.
func (a *LSAP[T]) Cost() T {
	return a.z
}

// DualPrices returns the dual prices associated with the assignment constraints
// of the U and V sets. If Assign is called prior, then complementary slackness
// conditions hold.
func (a *LSAP[T]) DualPrices() ap.DualPrices[T] {
	u := make([]T, a.n)
	v := make([]T, a.n)
	copy(u, a.u)
	copy(v, a.v)
	return ap.DualPrices[T]{U: u, V: v}
}

// M returns a large value used for removing edges.
func (a *LSAP[T]) M() T {
	return a.m
}

// ReducedCost returns the reduced cost of an edge. If Assign is called prior,
// then the reduced cost of a nonbasic (not in the optimal assignment) edge is
// the additional cost introducing that edge would incur.
func (a *LSAP[T]) ReducedCost(u, v int) T {
	return a.a[u][v] - a.u[u] - a.v[v]
}

// Remove takes an edge out of the solution. The edge's cost is set to the M
// value of the Assigner to discourage its use in subsequent calls to Assign.
//
// The solution algorithm provides an incremental primal-dual assignment problem
// solver. The initial optimization runs in O(n^3) time, after which one can
// remove edges from the resulting assignment and re-solve in O(n^2) time.
//
// For example, given some valid primal bound p:
//
//	if cost + rc > p {
//	    // (0,1) cannot be in an optimal solution
//	}
func (a *LSAP[T]) Remove(i, j int) {
	if a.f[i] == j {
		a.f[i] = -1
		a.fBar[j] = -1
	}
	a.a[i][j] = a.m
}

// SetM sets a large value used for removing edges.
func (a *LSAP[T]) SetM(m T) {
	a.m = m
}
