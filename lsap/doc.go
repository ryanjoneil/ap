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
// The dense solver is a port of the Fortran code APC for the dense AP given
// in:
//
//	Giorgio Carpaneto, Silvano Martello, and Paolo Toth.
//	"Algorithms and codes for the assignment problem."
//	In B. Simeone, P. Toth, G. Gallo, F. Maffioli, and S. Pallottino,
//	editors, "Fortran Codes for Network Optimization."
//	Annals of Operations Research 13 (1988): 193-223. Baltzer, Basel.
//
// APC implements Lawler's O(n³) version of the Hungarian algorithm, initialized
// with three edge preprocessing. Both are described in:
//
//	Rainer Burkard, Mauro Dell'Amico, and Silvano Martello.
//	"Assignment Problems - Revised Reprint."
//	Society for Industrial and Applied Mathematics (2012).
//
// See §4.2.2 (Algorithm 4.4, Augment, and Algorithm 4.5, Hungarian_3) for the
// solve, §4.4.4 (Algorithm 4.11, Three_edge_preprocessing) for the
// initialization, and §4.9 for APC. The book's bibliography lists the
// Carpaneto, Martello, and Toth paper as reference [169].
package lsap
