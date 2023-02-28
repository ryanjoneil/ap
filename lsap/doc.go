// Package lsap solves linear sum assignment problems. These take the form:
//
//	min   ∑_i,j c_ij * x_ij
//	s.t.  ∑_i   x_ij = 1      ∀ j
//	      ∑_j   x_ij = 1      ∀ i
//	            x_ij ∈ {0,1}  ∀ i,j
//
// x is an n-by-n matrix of binary variables representing edges in a bipartite
// graph, while c is the cost matrix for those edges. The constraints specify
// that each row (u ∈ U) is assigned to one column (v ∈ V) and vice versa.
//
//	a := lsap.New([][]int64{
//	    {10, 15, 12},
//	    {51, 75, 23},
//	    {11, 91, 10},
//	})
//
//	permutation := p.Assign() // [1 2 0]
//	cost := a.Cost()          // 49
//
// The LSAP struct conforms to ap.Int64DualPricer and ap.Int64ReducedCoster
// interfaces. Dual bounds and reduced costs are calculated as part of the
// assignment optimization. A copy method allows one to integrate an LSAP solver
// into a search routine for reduced cost-based domain filtering.
//
//	duals := a.DualPrices()   // {[0 14 1] [10 15 9]}
//	rc := a.ReducedCost(1, 1) // 46
//
// Given some valid primal bound p:
//
//	if cost + rc > p {
//	    // (0,1) cannot be in an optimal solution
//	}
//
// The solution algorithm provides an incremental primal-dual assignment problem
// solver. The initial optimization runs in O(n^3) time, after which one can
// remove edges from the resulting assignment and re-solve in O(n^2) time.
//
//	a.Remove(0, 1)
//	permutation = a.Assign() // [0 1 2]
//	cost = a.Cost()          // 95
//
// It is a port of the original Fortran code for the dense AP discussed in:
//
//	Giorgio Carpaneto, Silvano Martello, and Paolo Toth.
//	"Algorithms and codes for the assignment problem."
//	Annals of Operations Research 13, no. 1 (1988): 191-223.
package lsap
