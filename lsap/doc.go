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
// There are two solvers. [New] builds a dense solver from a complete n×n cost
// matrix, and [NewSparse] builds a sparse solver from a list of edges, where
// a missing edge is forbidden. Both support Remove, incremental re-solve and
// Copy, and both report dual prices and reduced costs.
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
//
// The sparse solver is the shortest path implementation of the Hungarian
// algorithm in §4.4 of the same book (§4.4.2, Algorithm 4.9, Procedure
// Shortest_path(k), and Algorithm 4.10, Hungarian_SP), using a binary heap
// for the Dijkstra search over the adjacency lists of labelled rows. It is
// initialized with the column-first variant of Basic_preprocessing (§4.1.2,
// Algorithm 4.1) restricted to existing edges. The book's §4.4.4 and §4.9.1
// describe the related sparse code LAPJVsp of
//
//	Roy Jonker and Anton Volgenant.
//	"A shortest augmenting path algorithm for dense and sparse linear
//	assignment problems."
//	Computing 38 (1987): 325-340.
//
// listed in the book's bibliography as reference [408]. The sparse solver
// does not use its preprocessing (§4.4.4, Algorithms 4.12 and 4.13).
package lsap
