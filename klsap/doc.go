// Package klsap ranks the k best assignments of a linear sum assignment
// problem (k-best LSAP, hence klsap). Given the LSAP
//
//	min   ∑_i,j c_ij * x_ij
//	s.t.  ∑_i   x_ij = 1      ∀ j
//	      ∑_j   x_ij = 1      ∀ i
//	            x_ij ∈ {0,1}  ∀ i,j
//
// and a count k, it finds k distinct assignments M_1, M_2, ..., M_k with
// z(M_1) ≤ z(M_2) ≤ ... ≤ z(M_k), where z(M) is the cost of M, such that
// every other assignment M has z(M) ≥ z(M_k). If the problem has fewer than
// k assignments, it ranks all of them.
//
// [New] builds a ranking from a complete n×n cost matrix, and [NewSparse]
// from a list of edges, where a missing edge is forbidden. Each call to Next
// returns the next assignment in nondecreasing order of cost.
//
// The ranking is Murty's method, from:
//
//	Katta G. Murty.
//	"An algorithm for ranking all the assignments in order of increasing
//	cost."
//	Operations Research 16(3) (1968): 682-687.
//
// It is described in:
//
//	Rainer Burkard, Mauro Dell'Amico, and Silvano Martello.
//	"Assignment Problems - Revised Reprint."
//	Society for Industrial and Applied Mathematics (2012).
//
// See §5.4.1 (Algorithm 5.1, Murty). The book's bibliography lists Murty's
// paper as reference [528]. Each subproblem is solved with package lsap; see
// its documentation for the sources of its dense and sparse solvers.
package klsap
