// Package graph builds the compressed-sparse-row bipartite graphs shared by
// the sparse lsap, mcm and lbap solvers.
//
// Every builder validates its input, copies it (the result never aliases
// caller memory) and runs in O(edges + NU + NV) time using two stable counting
// sorts (by column, then by row), so rows come out sorted by column without a
// comparison sort.
package graph

import (
	"errors"
	"fmt"
	"sort"

	"github.com/ryanjoneil/ap"
)

// CSR is a bipartite graph in compressed sparse row form. Row u's columns are
// Adj[Start[u]:Start[u+1]], sorted ascending.
type CSR struct {
	NU, NV int
	Start  []int // len NU+1
	Adj    []int // len = number of edges
}

var errEmpty = errors.New("graph: empty graph")

// build constructs the CSR for m edges whose endpoints are given by uv. It
// also returns order, where order[k] is the index of the input edge stored at
// Adj[k]. It reports out-of-range endpoints and duplicate edges.
func build(nU, nV, m int, uv func(i int) (int, int)) (*CSR, []int, error) {
	if nU <= 0 || nV <= 0 {
		return nil, nil, errEmpty
	}
	// Pass 1: stable counting sort of edge indices by column.
	colStart := make([]int, nV+1)
	for i := 0; i < m; i++ {
		u, v := uv(i)
		if u < 0 || u >= nU || v < 0 || v >= nV {
			return nil, nil, fmt.Errorf("graph: edge (%d, %d) out of range for %d x %d graph", u, v, nU, nV)
		}
		colStart[v+1]++
	}
	for v := 0; v < nV; v++ {
		colStart[v+1] += colStart[v]
	}
	byCol := make([]int, m)
	for i := 0; i < m; i++ {
		_, v := uv(i)
		byCol[colStart[v]] = i
		colStart[v]++
	}
	// Pass 2: stable counting sort by row; columns stay ascending within a row.
	start := make([]int, nU+1)
	for i := 0; i < m; i++ {
		u, _ := uv(i)
		start[u+1]++
	}
	for u := 0; u < nU; u++ {
		start[u+1] += start[u]
	}
	next := make([]int, nU)
	copy(next, start[:nU])
	adj := make([]int, m)
	order := make([]int, m)
	for _, i := range byCol {
		u, v := uv(i)
		adj[next[u]] = v
		order[next[u]] = i
		next[u]++
	}
	for u := 0; u < nU; u++ {
		for k := start[u] + 1; k < start[u+1]; k++ {
			if adj[k] == adj[k-1] {
				return nil, nil, fmt.Errorf("graph: duplicate edge (%d, %d)", u, adj[k])
			}
		}
	}
	return &CSR{NU: nU, NV: nV, Start: start, Adj: adj}, order, nil
}

// FromPairs builds an nU x nV graph from unweighted edges. It reports an error
// for an empty graph, an out-of-range endpoint or a duplicate edge.
func FromPairs(nU, nV int, pairs []ap.Pair) (*CSR, error) {
	g, _, err := build(nU, nV, len(pairs), func(i int) (int, int) { return pairs[i].U, pairs[i].V })
	return g, err
}

// FromAdjacency builds a graph from a boolean adjacency matrix with
// NU = len(adj) and NV = len(adj[0]). It reports an error for an empty or
// ragged matrix.
func FromAdjacency(adj [][]bool) (*CSR, error) {
	if len(adj) == 0 || len(adj[0]) == 0 {
		return nil, errEmpty
	}
	nU, nV := len(adj), len(adj[0])
	start := make([]int, nU+1)
	for u, row := range adj {
		if len(row) != nV {
			return nil, fmt.Errorf("graph: ragged adjacency matrix: row %d has %d columns, want %d", u, len(row), nV)
		}
		k := start[u]
		for _, ok := range row {
			if ok {
				k++
			}
		}
		start[u+1] = k
	}
	cols := make([]int, start[nU])
	for u, row := range adj {
		k := start[u]
		for v, ok := range row {
			if ok {
				cols[k] = v
				k++
			}
		}
	}
	return &CSR{NU: nU, NV: nV, Start: start, Adj: cols}, nil
}

// FromEdges builds an n x n graph from weighted edges. The returned costs are
// parallel to Adj. It reports an error for an empty graph, an out-of-range
// endpoint or a duplicate edge.
func FromEdges[T ap.Integer](n int, edges []ap.Edge[T]) (*CSR, []T, error) {
	g, order, err := build(n, n, len(edges), func(i int) (int, int) { return edges[i].U, edges[i].V })
	if err != nil {
		return nil, nil, err
	}
	costs := make([]T, len(order))
	for k, i := range order {
		costs[k] = edges[i].Cost
	}
	return g, costs, nil
}

// FromMatrix builds a complete n x n graph from a dense square cost matrix in
// which every entry is an edge. The returned costs are parallel to Adj. It
// reports an error for an empty or non-square matrix.
func FromMatrix[T ap.Integer](costs [][]T) (*CSR, []T, error) {
	n := len(costs)
	if n == 0 || len(costs[0]) == 0 {
		return nil, nil, errEmpty
	}
	flat := make([]T, 0, n*n)
	for u, row := range costs {
		if len(row) != n {
			return nil, nil, fmt.Errorf("graph: matrix is not square: row %d has %d columns, want %d", u, len(row), n)
		}
		flat = append(flat, row...)
	}
	start := make([]int, n+1)
	adj := make([]int, n*n)
	for u := 0; u < n; u++ {
		start[u+1] = (u + 1) * n
		row := adj[u*n : (u+1)*n]
		for v := range row {
			row[v] = v
		}
	}
	return &CSR{NU: n, NV: n, Start: start, Adj: adj}, flat, nil
}

// Find returns the index into Adj of edge (u, v), or -1 if the edge is absent
// or out of range.
func (g *CSR) Find(u, v int) int {
	if u < 0 || u >= g.NU {
		return -1
	}
	lo, hi := g.Start[u], g.Start[u+1]
	k := lo + sort.SearchInts(g.Adj[lo:hi], v)
	if k < hi && g.Adj[k] == v {
		return k
	}
	return -1
}
