// Package mcm solves maximum cardinality matching (MCM) problems on bipartite
// graphs. Given a graph with rows U and columns V, a matching is a set of edges, no
// two sharing a row or a column; the solver finds one with as many edges as
// possible. For a square graph this is a perfect matching (a feasible
// assignment) exactly when its size is |U| = |V|.
//
// [New] builds a solver from a boolean adjacency matrix, and [NewSparse]
// builds one from a list of edges. Both support Remove, incremental
// re-solve and Copy.
//
// The solver is the Hopcroft–Karp algorithm:
//
//	John E. Hopcroft and Richard M. Karp.
//	"An n^5/2 algorithm for maximum matchings in bipartite graphs."
//	SIAM Journal on Computing 2 (1973): 225–231.
//
// It is listed as reference [389] in the bibliography of
//
//	Rainer Burkard, Mauro Dell'Amico, and Silvano Martello.
//	"Assignment Problems - Revised Reprint."
//	Society for Industrial and Applied Mathematics (2012).
//
// which describes it in §3.3 (Algorithm 3.2, Hopcroft_Karp). The initial
// matching is greedy: each row, in increasing order, takes the first free
// column in column order among its edges. After a Remove, the solver
// re-optimizes from the previous matching rather than starting over.
package mcm
