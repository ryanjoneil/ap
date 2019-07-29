package lsap

import "github.com/ryanjoneil/ap"

// LSAP solves linear sum assignment problems.
type LSAP struct {
	A [][]int64 // A[i][j] = cost of assigning row i to column j
	U []int64   // U[i] = dual price for row i
	V []int64   // V[j] = dual price for column j
	Z int64     // Objective value

	f    []int   // f[i] = column assigned to row i, -1 if unassigned
	fBar []int   // fBar[j] = row assigned to column j, -1 if unassigned
	p    []int   // First column of row i not yet examined in phase 2
	c    []int   // c[j] = row preceding column j in current alternating path
	pi   []int64 // pi[j] = min { a[i,j] - u[i] - v[j] | i in lr, i != fBar[j] }

	M    int64
	Size int
}

// Assign returns an optimal assignment for an LSAP.
func (a *LSAP) Assign() ap.Assignment {
	a.solve()
	assign := make(ap.Assignment, a.Size)
	copy(assign, a.f)
	return assign
}

// Cost returns the objective value of an LSAP. If Assign is called prior to
// calling Cost, then the value is minimal.
func (a *LSAP) Cost() int64 {
	return a.Z
}

// DualPrices returns the dual prices associated with the assignment constraints
// of the U and V sets. If Assign is called prior, then complementary slackness
// conditions hold.
func (a *LSAP) DualPrices() ap.Int64DualPrices {
	u := make([]int64, a.Size)
	v := make([]int64, a.Size)
	copy(u, a.U)
	copy(v, a.V)
	return ap.Int64DualPrices{U: u, V: v}
}

// ReducedCost returns the reduced cost of an edge. If Assign is called prior,
// then the reduced cost of a nonbasic (not in the optimal assignment) edge is
// the additional cost introducing that edge would incur.
func (a *LSAP) ReducedCost(u, v int) int64 {
	return a.A[u][v] - a.U[u] - a.V[v]
}

// Remove takes an edge out of the solution. The edge's cost is set to the M
// value of the LSAP struct to discourage its use in subsequent calls to Assign.
func (a *LSAP) Remove(i, j int) {
	if a.f[i] == j {
		a.f[i] = -1
		a.fBar[j] = -1
	}
	a.A[i][j] = a.M
}
