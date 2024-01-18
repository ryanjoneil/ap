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
// It is a port of the original Fortran code for the dense AP discussed in:
//
//	Giorgio Carpaneto, Silvano Martello, and Paolo Toth.
//	"Algorithms and codes for the assignment problem."
//	Annals of Operations Research 13, no. 1 (1988): 191-223.
package lsap
