// Package lbap solves linear bottleneck assignment problems. These take the
// form:
//
//	min   max_i,j c_ij * x_ij
//	s.t.  ∑_i   x_ij = 1      ∀ j
//	      ∑_j   x_ij = 1      ∀ i
//	            x_ij ∈ {0,1}  ∀ i,j
//
// That is, among all perfect assignments of rows (u ∈ U) to columns (v ∈ V),
// find one whose largest assigned cost is as small as possible. Bottleneck
// problems usually have many optimal assignments; which one is returned is
// deterministic but otherwise unspecified.
//
// [New] builds a solver from a complete n×n cost matrix, and [NewSparse]
// builds one from a list of edges, where a missing edge is forbidden. Both
// support Remove, incremental re-solve and Copy.
//
// The solver is a threshold algorithm, an idea that dates back to
//
//	R. Garfinkel.
//	"An improved algorithm for the bottleneck assignment problem."
//	Operations Research 19 (1971): 1747–1751.
//
// listed as reference [313] in the bibliography of
//
//	Rainer Burkard, Mauro Dell'Amico, and Silvano Martello.
//	"Assignment Problems - Revised Reprint."
//	Society for Industrial and Applied Mathematics (2012).
//
// which describes it in §6.2.2 (Algorithm 6.1, Threshold). Each threshold
// c* is checked by asking whether the bipartite graph of edges with cost at
// most c* has a perfect matching, using the Hopcroft–Karp algorithm
//
//	John E. Hopcroft and Richard M. Karp.
//	"An n^5/2 algorithm for maximum matchings in bipartite graphs."
//	SIAM Journal on Computing 2 (1973): 225–231.
//
// (book reference [389], §3.3, Algorithm 3.2, Hopcroft_Karp).
//
// Departures from Algorithm 6.1:
//
//   - The search bisects the index range of the sorted distinct costs, rather
//     than testing the median of the cost entries strictly between the
//     current bounds. Each test halves the number of candidate values, so
//     there are O(log k) tests for k distinct costs. The largest cost is
//     tested explicitly, since sparse input or removals may leave no perfect
//     assignment at all.
//   - Each feasibility check warm-starts Hopcroft–Karp from the matching left
//     by the previous check, after unmatching the rows whose matched edge is
//     above the new threshold, instead of starting from an empty matching.
//   - After a Remove, the search starts at the previous optimum, since
//     removing edges never decreases the bottleneck value.
package lbap
