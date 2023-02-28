package lsap

import "github.com/ryanjoneil/ap"

type lsap struct {
	M int64 // A large cost to avoid using edges
	n int   // n of assignment problem

	a [][]int64 // a[i][j] = cost of assigning row i to column j
	u []int64   // u[i] = dual price for row i
	v []int64   // v[j] = dual price for column j
	z int64     // Objective value

	f    []int   // f[i] = column assigned to row i, -1 if unassigned
	fBar []int   // fBar[j] = row assigned to column j, -1 if unassigned
	p    []int   // First column of row i not yet examined in phase 2
	c    []int   // c[j] = row preceding column j in current alternating path
	pi   []int64 // pi[j] = min { a[i,j] - u[i] - v[j] | i in lr, i != fBar[j] }
}

// Assign returns an optimal assignment for an LSAP.
func (a *lsap) Assign() ap.Permutation {
	a.solve()
	p := make(ap.Permutation, a.n)
	copy(p, a.f)
	return p
}

// Cost returns the objective value of an LSAP. If Assign is called prior to
// calling Cost, then the value is minimal.
func (a *lsap) Cost() int64 {
	return a.z
}

// DualPrices returns the dual prices associated with the assignment constraints
// of the U and V sets. If Assign is called prior, then complementary slackness
// conditions hold.
func (a *lsap) DualPrices() ap.Int64DualPrices {
	u := make([]int64, a.n)
	v := make([]int64, a.n)
	copy(u, a.u)
	copy(v, a.v)
	return ap.Int64DualPrices{U: u, V: v}
}

// ReducedCost returns the reduced cost of an edge. If Assign is called prior,
// then the reduced cost of a nonbasic (not in the optimal assignment) edge is
// the additional cost introducing that edge would incur.
func (a *lsap) ReducedCost(u, v int) int64 {
	return a.a[u][v] - a.u[u] - a.v[v]
}

// Remove takes an edge out of the solution. The edge's cost is set to the M
// value of the LSAP struct to discourage its use in subsequent calls to Assign.
func (a *lsap) Remove(i, j int) {
	if a.f[i] == j {
		a.f[i] = -1
		a.fBar[j] = -1
	}
	a.a[i][j] = a.M
}
