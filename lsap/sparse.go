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
//   - The book does its arithmetic in unbounded integers. Here the kernel keeps
//     every dual price within the margin b, and stops a search at the first
//     label above the cap l, so that the unchecked labels and reduced costs
//     cannot overflow T (see margin). initialize checks each dual price it
//     sets, augment checks each popped label against l and each dual update
//     against b, and assign sums the objective value with checked additions
//     unless n·A fits in T. A failed check sets the sticky overflow flag, and
//     assign returns ap.ErrOverflow.
type sparse[T ap.Integer] struct {
	maxT T   // Largest value of T, cached for the search
	n    int // n of assignment problem

	b        T    // Margin: every |u[i]| and |v[j]| is at most b
	l        T    // Label cap: every label the search scans is at most l
	overflow bool // True once solving would overflow T; sticky
	sumFits  bool // True if n·A fits in T, so that no cost sum overflows
	z        T    // Objective value

	g       *graph.CSR // Edges; shared between copies and never modified
	c       []T        // c[e] = cost of edge e, parallel to g.Adj; shared
	removed []bool     // removed[e] = true if edge e is removed; nil if none

	u       []T   // u[i] = dual price for row i
	v       []T   // v[j] = dual price for column j
	rowMate []int // rowMate[i] = column assigned to row i, -1 if unassigned
	rowEdge []int // rowEdge[i] = edge (i, rowMate[i]), if rowMate[i] ≥ 0
	colMate []int // colMate[j] = row assigned to column j, -1 if unassigned

	// infeasible is set when some row has no augmenting path, or when a row or
	// column has no edges. Removals only shrink the graph, so it stays set.
	// It and overflow are never both set.
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
// and initializes it. See NewSparse for the errors it returns. If
// initializing would overflow T, newSparse still succeeds, and assign returns
// ap.ErrOverflow.
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

	maxT, minT := intmath.MaxOf[T](), intmath.MinOf[T]()
	var bigA T // max |c[e]| over all edges
	for _, ce := range c {
		if ce == minT || ce == maxT {
			return nil, fmt.Errorf("lsap: cost %d is out of range for %T", ce, ce)
		}
		bigA = max(bigA, abs(ce))
	}

	a := &sparse[T]{
		maxT:    maxT,
		n:       n,
		g:       g,
		c:       c,
		u:       make([]T, n),
		v:       make([]T, n),
		rowMate: make([]int, n),
		rowEdge: make([]int, n),
		colMate: make([]int, n),
	}
	a.b, a.l = a.margin(bigA)
	_, a.sumFits = intmath.MulInt(bigA, n)
	a.initialize()
	return a, nil
}

// margin returns the margin b on the dual prices and the cap l on the labels
// the search scans, given the largest cost magnitude A, as
//
//	b = (maxT − A − 1) / 4   and   l = maxT − A − 2b − 1.
//
// Then A + 2b ≤ maxT − 1 − 2b, and l ≥ 2b ≥ 0. Edges are never modified, and
// removing one only marks it, so A, b and l never change.
//
// If |u[i]| ≤ b and |v[j]| ≤ b, then c[e] − u[i] and c[e] − v[j] lie in
// [−(A+b), A+b], and c[e] − u[i] − v[j] in [−(A+2b), A+2b], so none of the
// reduced costs that initialize and reducedCost compute overflows.
//
// The search labels the columns adjacent to s with reduced costs, which are
// at most A + 2b ≤ maxT − 1. When it scans column j with label d, it labels
// each column k adjacent to row i = colMate[j] with (d − u[i]) + c[e] − v[k].
// Dual feasibility makes every reduced cost, and so every label, nonnegative.
// If 0 ≤ d ≤ l, then d − u[i] lies in [−b, l + b], adding c[e] gives
// [−(A+b), l + A + b], and subtracting v[k] gives [−(A+2b), l + A + 2b] =
// [−(A+2b), maxT − 1]. So no intermediate overflows, and every label is at
// most maxT − 1, below the maxT that marks an unlabelled column.
//
// The cap costs no false alarms. The search scans labels in nondecreasing
// order and ends with δ, the sink's label, so if it scans a label above
// l ≥ 2b, then either it finds no sink or δ > 2b. In the second case
// u[s] + δ > b, since u[s] ≥ −b, and the dual update would fail anyway.
//
// The dual prices change only in initialize, which checks each one against b
// before using it, and in the dual update, which raises u[s] by δ and, for
// each scanned column j, raises u[colMate[j]] and lowers v[j] by r = δ −
// dist[j]. Both δ and r lie in [0, l], since 0 ≤ dist[j] ≤ δ ≤ l. So the
// update checks only δ ≤ b − u[s], r ≤ b − u[colMate[j]] and r ≤ v[j] + b,
// whose right-hand sides lie in [0, 2b] and can't overflow.
//
// No cost is the smallest value of T, so A ≤ maxT and maxT − A − 1 doesn't
// overflow. If A = maxT, then there is no margin, and margin returns −1 for
// both, so that every check fails.
func (a *sparse[T]) margin(bigA T) (b, l T) {
	if bigA == a.maxT {
		return -1, -1
	}
	b = (a.maxT - bigA - 1) / 4
	return b, a.maxT - bigA - 2*b - 1
}

// inMargin reports whether |x| ≤ a.b, without computing |x|.
func (a *sparse[T]) inMargin(x T) bool {
	return -a.b <= x && x <= a.b
}

// initialize performs column reduction, then row reduction, then a greedy
// partial assignment on zero reduced cost edges. It marks the problem
// infeasible if any row or column has no edges, and does no arithmetic in
// that case, so that infeasibility takes precedence over overflow. Otherwise
// it checks each dual price against the margin (see margin) as it sets it,
// before the price is used. If one is outside the margin, initialize sets
// a.overflow and stops.
func (a *sparse[T]) initialize() {
	g, c, u, v := a.g, a.c, a.u, a.v

	for i := 0; i < a.n; i++ {
		a.rowMate[i] = -1
		a.colMate[i] = -1
		if g.Start[i] == g.Start[i+1] {
			a.infeasible = true
		}
	}
	seen := make([]bool, a.n) // seen[j] = true if column j has an edge
	for e, j := range g.Adj {
		if !seen[j] {
			seen[j] = true
			v[j] = c[e]
		} else {
			v[j] = min(v[j], c[e])
		}
	}
	for j := range v {
		if !seen[j] {
			a.infeasible = true
		}
	}
	if a.infeasible {
		return
	}

	for j := range v {
		if !a.inMargin(v[j]) {
			a.overflow = true
			return
		}
	}

	for i := 0; i < a.n; i++ {
		lo, hi := g.Start[i], g.Start[i+1]
		ui := c[lo] - v[g.Adj[lo]]
		for e := lo + 1; e < hi; e++ {
			ui = min(ui, c[e]-v[g.Adj[e]])
		}
		if !a.inMargin(ui) {
			a.overflow = true
			return
		}
		u[i] = ui
		// Columns are sorted within a row, so this takes the lowest index.
		for e := lo; e < hi; e++ {
			j := g.Adj[e]
			if a.colMate[j] < 0 && c[e]-ui-v[j] == 0 {
				a.rowMate[i] = j
				a.rowEdge[i] = e
				a.colMate[j] = i
				break
			}
		}
	}
}

// assign solves and sums the cost of the assignment. Every partial sum has at
// most n terms, each of magnitude at most A, so if n·A fits in T, no partial
// sum overflows, and the additions go unchecked. Otherwise they are checked.
func (a *sparse[T]) assign() (ap.Permutation, error) {
	if a.infeasible {
		return nil, ap.ErrInfeasible
	}
	if a.overflow {
		return nil, ap.ErrOverflow
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
			if a.overflow {
				return nil, ap.ErrOverflow
			}
			a.infeasible = true
			return nil, ap.ErrInfeasible
		}
	}

	var z T
	for _, e := range a.rowEdge {
		ce := a.c[e]
		if a.sumFits {
			z += ce
			continue
		}
		var ok bool
		if z, ok = intmath.Add(z, ce); !ok {
			a.overflow = true
			return nil, ap.ErrOverflow
		}
	}
	a.z = z

	p := make(ap.Permutation, a.n)
	copy(p, a.rowMate)
	return p, nil
}

// augment finds a shortest augmenting path from the unassigned row s over
// reduced costs, updates the duals and augments the assignment along the
// path. This is Procedure Shortest_path(s) (book Algorithm 4.9) followed by
// the augmentation of Hungarian_SP (Algorithm 4.10). It returns false, and
// changes nothing, if no unassigned column can be reached from s.
//
// As a departure from the source, augment stops at the first scanned label
// above the cap a.l, and checks that the dual update keeps the dual prices
// within the margin a.b (see margin), which makes the unchecked labels safe.
// If either check fails, augment sets a.overflow and returns false, without
// augmenting. It may then have updated some of the dual prices, but each
// stays within the margin.
func (a *sparse[T]) augment(s int) bool {
	g, c, u, v := a.g, a.c, a.u, a.v
	dist, pred, done := a.dist, a.pred, a.done
	removed, colMate := a.removed, a.colMate
	maxT, l := a.maxT, a.l
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
		if it.d > l {
			a.overflow = true
			break
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

	if sink >= 0 && !a.updateDuals(s, delta, scanned) {
		a.overflow = true
		sink = -1
	}

	if sink >= 0 {
		// Augment along the path from sink back to s.
		for j := sink; ; {
			i := pred[j]
			a.colMate[j] = i
			next := a.rowMate[i]
			a.rowMate[i] = j
			a.rowEdge[i] = g.Find(i, j)
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

// updateDuals raises u[s] by δ and, for each scanned column j, raises
// u[colMate[j]] and lowers v[j] by r = δ − dist[j]. This is the dual update
// of book Algorithm 4.9, with the departures listed on sparse. It returns
// false, leaving the remaining prices unchanged, at the first update that
// would leave the margin a.b. 0 ≤ r ≤ δ, so no check can overflow: b − u[i]
// and v[j] + b lie in [0, 2b].
func (a *sparse[T]) updateDuals(s int, delta T, scanned []int) bool {
	u, v, b := a.u, a.v, a.b
	if delta > b-u[s] {
		return false
	}
	u[s] += delta
	for _, j := range scanned {
		r := delta - a.dist[j]
		i := a.colMate[j]
		if r > b-u[i] || r > v[j]+b {
			return false
		}
		u[i] += r
		v[j] -= r
	}
	return true
}

// cost returns the objective value, which assign sums with checked additions.
func (a *sparse[T]) cost() T {
	return a.z
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
		b:          a.b,
		l:          a.l,
		overflow:   a.overflow,
		sumFits:    a.sumFits,
		z:          a.z,
		g:          a.g,
		c:          a.c,
		removed:    removed,
		u:          append([]T(nil), a.u...),
		v:          append([]T(nil), a.v...),
		rowMate:    append([]int(nil), a.rowMate...),
		rowEdge:    append([]int(nil), a.rowEdge...),
		colMate:    append([]int(nil), a.colMate...),
		infeasible: a.infeasible,
	}
}
