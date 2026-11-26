package lsap

import (
	"fmt"
	"math"

	"github.com/ryanjoneil/ap"
	"github.com/ryanjoneil/ap/internal/graph"
	"github.com/ryanjoneil/ap/internal/intmath"
)

// sparse is the kernel for a sparse edge list. It is the shortest path
// implementation of the Hungarian algorithm, Hungarian_SP (book §4.4.2,
// Algorithm 4.10), which calls Procedure Shortest_path(k) (Algorithm 4.9) for
// each unassigned row k. The search runs only over the adjacency lists of
// labelled rows, so it never touches a missing edge.
//
// This kernel follows the book's algorithms, not the code LAPJVsp of Jonker
// and Volgenant (1987), book ref [408], which also works on a sparse forward
// star. In particular it does not use their preprocessing (book §4.4.4,
// Algorithms 4.12 Reduction_transfer and 4.13 Augmenting_row_reduction).
//
// The initial duals and partial assignment come from Basic_preprocessing (book
// §4.1.2, Algorithm 4.1) restricted to existing edges, in the column-first
// order the book mentions as an alternative: column reduction, then row
// reduction, then each row takes its lowest free column of zero reduced cost.
//
// Departures from Algorithm 4.9:
//   - The argmin over unscanned columns uses a binary heap with lazy deletion
//     (a label decrease pushes a new entry, and entries popped for scanned
//     columns are skipped) instead of a pass over all of V, so a search costs
//     O(m log m) for m edges rather than O(n²).
//   - Ties in the argmin are broken by lowest column index, so results are
//     deterministic.
//   - Relaxing edge (i, k) does not test whether column k is scanned. A
//     scanned column has a label no larger than the current one, and reduced
//     costs are nonnegative, so the label comparison already rejects it.
//   - The dual update skips the sink, whose change δ − π_sink is zero, and
//     visits only scanned columns.
//   - Labels are reset only for the columns a search touched.
//   - Edges can be removed after solving, and the problem re-solved
//     incrementally from the current duals and partial assignment. Removing
//     an edge keeps the duals feasible; removing an assigned edge frees its
//     row and column for the next search.
type sparse[T ap.Integer] struct {
	maxT T   // Largest value of T, cached for the search
	n    int // n of assignment problem

	g       *graph.CSR // Edges; shared between copies and never modified
	c       []T        // c[e] = cost of edge e, parallel to g.Adj; shared
	removed []bool     // removed[e] = true if edge e is removed; nil if none

	u       []T   // u[i] = dual price for row i
	v       []T   // v[j] = dual price for column j
	rowMate []int // rowMate[i] = column assigned to row i, -1 if unassigned
	colMate []int // colMate[j] = row assigned to column j, -1 if unassigned

	// infeasible is set when some row has no augmenting path, or when a row or
	// column has no edges. Removals only shrink the graph, so it stays set.
	infeasible bool

	// Scratch space for augment, allocated on the first call to assign. Between
	// searches, dist[j] == maxT and done[j] == false for every column j.
	dist    []T        // dist[j] = shortest path label of column j (π_j)
	pred    []int      // pred[j] = row preceding column j on its path
	done    []bool     // done[j] = true if column j is scanned (j ∈ SV)
	touched []int      // Columns whose labels were set by this search
	scanned []int      // Columns scanned by this search, in order
	heap    minHeap[T] // Unscanned columns by (dist, column)
}

// newSparse builds a sparse kernel for an n×n problem with the given edges
// and initializes it. See NewSparse for the errors it returns.
func newSparse[T ap.Integer](n int, edges []ap.Edge[T]) (*sparse[T], error) {
	if n < 1 {
		return nil, fmt.Errorf("lsap: n = %d; want n ≥ 1", n)
	}
	if n > math.MaxInt32 {
		return nil, fmt.Errorf("lsap: n = %d is larger than the limit of %d", n, math.MaxInt32)
	}

	g, c, err := graph.FromEdges(n, edges)
	if err != nil {
		return nil, fmt.Errorf("lsap: %w", err)
	}

	if len(c) > 0 {
		minC, maxC := c[0], c[0]
		for _, ce := range c {
			minC = min(minC, ce)
			maxC = max(maxC, ce)
		}
		k := 4 * (n + 1)
		if !intmath.FitsScaled(k, minC) || !intmath.FitsScaled(k, maxC) {
			return nil, fmt.Errorf(
				"lsap: cost range [%d, %d] could overflow %T for n = %d",
				minC, maxC, minC, n,
			)
		}
	}

	a := &sparse[T]{
		maxT:    intmath.MaxOf[T](),
		n:       n,
		g:       g,
		c:       c,
		u:       make([]T, n),
		v:       make([]T, n),
		rowMate: make([]int, n),
		colMate: make([]int, n),
	}
	a.initialize()
	return a, nil
}

// initialize performs column reduction, then row reduction, then a greedy
// partial assignment on zero reduced cost edges. It marks the problem
// infeasible if any row or column has no edges.
func (a *sparse[T]) initialize() {
	g, c, u, v := a.g, a.c, a.u, a.v

	for j := range v {
		v[j] = a.maxT
		a.colMate[j] = -1
	}
	for e, j := range g.Adj {
		v[j] = min(v[j], c[e])
	}
	for j := range v {
		// Every cost is below maxT by the overflow bound, so v[j] == maxT
		// means column j has no edges.
		if v[j] == a.maxT {
			v[j] = 0
			a.infeasible = true
		}
	}

	for i := 0; i < a.n; i++ {
		a.rowMate[i] = -1
		lo, hi := g.Start[i], g.Start[i+1]
		if lo == hi {
			a.infeasible = true
			continue
		}
		ui := c[lo] - v[g.Adj[lo]]
		for e := lo + 1; e < hi; e++ {
			ui = min(ui, c[e]-v[g.Adj[e]])
		}
		u[i] = ui
		// Columns are sorted within a row, so this takes the lowest index.
		for e := lo; e < hi; e++ {
			j := g.Adj[e]
			if a.colMate[j] < 0 && c[e]-ui-v[j] == 0 {
				a.rowMate[i] = j
				a.colMate[j] = i
				break
			}
		}
	}
}

func (a *sparse[T]) assign() (ap.Permutation, error) {
	if a.infeasible {
		return nil, ap.ErrInfeasible
	}
	if a.dist == nil {
		a.dist = make([]T, a.n)
		a.pred = make([]int, a.n)
		a.done = make([]bool, a.n)
		for j := range a.dist {
			a.dist[j] = a.maxT
		}
	}
	for s := 0; s < a.n; s++ {
		if a.rowMate[s] < 0 && !a.augment(s) {
			a.infeasible = true
			return nil, ap.ErrInfeasible
		}
	}
	p := make(ap.Permutation, a.n)
	copy(p, a.rowMate)
	return p, nil
}

// augment finds a shortest augmenting path from the unassigned row s over
// reduced costs, updates the duals and augments the assignment along the
// path. This is Procedure Shortest_path(s) (book Algorithm 4.9) followed by
// the augmentation of Hungarian_SP (Algorithm 4.10). It returns false, and
// changes nothing, if no unassigned column can be reached from s.
func (a *sparse[T]) augment(s int) bool {
	g, c, u, v := a.g, a.c, a.u, a.v
	dist, pred, done := a.dist, a.pred, a.done
	removed, colMate := a.removed, a.colMate
	maxT := a.maxT
	h := a.heap[:0]
	touched := a.touched[:0]
	scanned := a.scanned[:0]

	// Label the columns adjacent to s. There are no duplicate edges, so each
	// column is labelled once here.
	us := u[s]
	for e := g.Start[s]; e < g.Start[s+1]; e++ {
		if removed != nil && removed[e] {
			continue
		}
		j := g.Adj[e]
		d := c[e] - us - v[j]
		dist[j] = d
		pred[j] = s
		touched = append(touched, j)
		h.push(heapItem[T]{d: d, j: int32(j)})
	}

	sink := -1
	var delta T // Length of the shortest augmenting path (δ)
	for h.len() > 0 {
		it := h.pop()
		j := int(it.j)
		if done[j] {
			// Stale entry. A column's label only decreases, so its current
			// entry is popped before any stale one.
			continue
		}
		done[j] = true
		if colMate[j] < 0 {
			sink, delta = j, it.d
			break
		}
		scanned = append(scanned, j)

		// Extend the path through row i = colMate[j], whose label is dist[j].
		i := colMate[j]
		d := it.d - u[i]
		for e := g.Start[i]; e < g.Start[i+1]; e++ {
			if removed != nil && removed[e] {
				continue
			}
			// Departure: no test of done[k]. A scanned column k has
			// dist[k] ≤ dist[j] ≤ nd, so the comparison rejects it.
			k := g.Adj[e]
			if nd := d + c[e] - v[k]; nd < dist[k] {
				if dist[k] == maxT {
					touched = append(touched, k)
				}
				dist[k] = nd
				pred[k] = i
				h.push(heapItem[T]{d: nd, j: int32(k)})
			}
		}
	}

	if sink >= 0 {
		// Dual update.
		u[s] += delta
		for _, j := range scanned {
			r := delta - dist[j]
			u[colMate[j]] += r
			v[j] -= r
		}

		// Augment along the path from sink back to s.
		for j := sink; ; {
			i := pred[j]
			a.colMate[j] = i
			next := a.rowMate[i]
			a.rowMate[i] = j
			if i == s {
				break
			}
			j = next
		}
	}

	for _, j := range touched {
		dist[j] = maxT
		done[j] = false
	}
	a.heap = h[:0]
	a.touched = touched[:0]
	a.scanned = scanned[:0]
	return sink >= 0
}

// cost sums the costs of the assigned edges. The overflow bound guarantees
// that n·max|c| fits in T.
func (a *sparse[T]) cost() T {
	var z T
	for i, j := range a.rowMate {
		if e := a.g.Find(i, j); e >= 0 {
			z += a.c[e]
		}
	}
	return z
}

func (a *sparse[T]) dualPrices() ap.DualPrices[T] {
	u := make([]T, a.n)
	v := make([]T, a.n)
	copy(u, a.u)
	copy(v, a.v)
	return ap.DualPrices[T]{U: u, V: v}
}

// live returns the index of edge (i, j), or -1 if it is out of range, missing
// or removed.
func (a *sparse[T]) live(i, j int) int {
	e := a.g.Find(i, j)
	if e < 0 || (a.removed != nil && a.removed[e]) {
		return -1
	}
	return e
}

func (a *sparse[T]) reducedCost(i, j int) (T, bool) {
	e := a.live(i, j)
	if e < 0 {
		return 0, false
	}
	return a.c[e] - a.u[i] - a.v[j], true
}

// remove marks edge (i, j) removed. If the edge is assigned, it is unassigned,
// so the next solve re-optimizes row i.
func (a *sparse[T]) remove(i, j int) {
	e := a.live(i, j)
	if e < 0 {
		return
	}
	if a.removed == nil {
		a.removed = make([]bool, len(a.c))
	}
	a.removed[e] = true
	if a.rowMate[i] == j {
		a.rowMate[i] = -1
		a.colMate[j] = -1
	}
}

// copy shares the graph and costs, which are never modified, and clones the
// removals and the primal and dual solution. Scratch space is allocated by
// the copy's first call to assign.
func (a *sparse[T]) copy() kernel[T] {
	var removed []bool
	if a.removed != nil {
		removed = append([]bool(nil), a.removed...)
	}
	return &sparse[T]{
		maxT:       a.maxT,
		n:          a.n,
		g:          a.g,
		c:          a.c,
		removed:    removed,
		u:          append([]T(nil), a.u...),
		v:          append([]T(nil), a.v...),
		rowMate:    append([]int(nil), a.rowMate...),
		colMate:    append([]int(nil), a.colMate...),
		infeasible: a.infeasible,
	}
}
