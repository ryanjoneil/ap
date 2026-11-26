package lsap

import (
	"fmt"
	"sync/atomic"

	"github.com/ryanjoneil/ap"
	"github.com/ryanjoneil/ap/internal/intmath"
)

// dense is the kernel for a complete cost matrix. It is a port of the Fortran
// code APC of Carpaneto, Martello and Toth (1988): Lawler's O(n³) Hungarian
// algorithm (book §4.2.2, Algorithms 4.4 and 4.5), initialized with three edge
// preprocessing (book §4.4.4, Algorithm 4.11). See initialize.go and solve.go.
//
// Departure from the source: APC solves a single problem. Here an edge can be
// removed after solving by setting its cost to the big-M value m, and the
// problem re-solved incrementally from the current primal and dual solution.
// The cost rows are copied on write, so that Copy is cheap. Copy marks the
// source's rows as shared, so owned is atomic: several goroutines may call
// Copy on the same kernel at once. Only copy and remove touch owned.
//
// Departure from the source: APC does its arithmetic unchecked, in a type
// assumed wide enough. Here the kernel keeps every dual price within the
// margin b (see margin), so that the reduced costs in the inner loop of path
// cannot overflow T. initialize checks each dual price it sets, path checks
// each dual update, and assign rechecks all of them against the current b.
// A dual price outside b sets the sticky overflow flag, and assign returns
// ap.ErrOverflow. The objective value is summed with checked additions.
type dense[T ap.Integer] struct {
	m    T   // Cost of a removed edge; every real cost is below m
	maxT T   // Largest value of T, cached for path
	n    int // n of assignment problem

	a0       T    // max |a[i][j]| over the original costs
	b        T    // Margin: every |u[i]| and |v[j]| is at most b
	removed  bool // True once an edge has been removed, so that m is a cost
	overflow bool // True once solving would overflow T; sticky

	a     [][]T         // a[i][j] = cost of assigning row i to column j
	owned []atomic.Bool // owned[i] = true if a[i] is not shared with another kernel
	u     []T           // u[i] = dual price for row i
	v     []T           // v[j] = dual price for column j
	z     T             // Objective value

	f    []int // f[i] = column assigned to row i, -1 if unassigned
	fBar []int // fBar[j] = row assigned to column j, -1 if unassigned
	p    []int // First column of row i not yet examined in phase 2
	c    []int // c[j] = row preceding column j in current alternating path
	pi   []T   // pi[j] = min { a[i,j] - u[i] - v[j] | i in lr, i != fBar[j] }
}

// newDense copies the square, non-empty cost matrix costs into a dense kernel
// and initializes it. It returns an error if a cost is the smallest value of
// T, or if M doesn't fit below the largest value of T; see New. If
// initializing would overflow T, newDense still succeeds, and assign returns
// ap.ErrOverflow.
func newDense[T ap.Integer](costs [][]T) (*dense[T], error) {
	n := len(costs)

	minC, maxC := costs[0][0], costs[0][0]
	for _, row := range costs {
		for _, cij := range row {
			minC = min(minC, cij)
			maxC = max(maxC, cij)
		}
	}

	maxT := intmath.MaxOf[T]()
	m, ok := bigM(n, minC, maxC)
	if !ok || m == maxT || minC == intmath.MinOf[T]() {
		return nil, fmt.Errorf(
			"lsap: cost range [%d, %d] is too wide for %T with n = %d",
			minC, maxC, minC, n,
		)
	}

	a := make([][]T, n)
	owned := make([]atomic.Bool, n)
	for i, row := range costs {
		a[i] = append([]T(nil), row...)
		owned[i].Store(true)
	}

	f := make([]int, n)
	fBar := make([]int, n)
	p := make([]int, n)
	c := make([]int, n)
	pi := make([]T, n)

	for i := 0; i < n; i++ {
		f[i] = -1
		fBar[i] = -1
		p[i] = -1
		c[i] = -1
		pi[i] = -1
	}

	d := &dense[T]{
		m:     m,
		maxT:  maxT,
		n:     n,
		a0:    max(abs(minC), abs(maxC)),
		a:     a,
		owned: owned,
		u:     make([]T, n),
		v:     make([]T, n),
		f:     f,
		fBar:  fBar,
		p:     p,
		c:     c,
		pi:    pi,
	}
	d.b = d.margin()
	d.initialize()
	return d, nil
}

// abs returns |x|. x must not be the smallest value of T.
func abs[T ap.Integer](x T) T {
	if x < 0 {
		return -x
	}
	return x
}

// margin returns the bound b on the dual prices for the current costs, or −1
// if there is none. Let A be the largest |a[i][j]|: a0 until the first
// removal, then max(a0, |m|). Then
//
//	b = (maxT − A − 1) / 2.
//
// If |u[i]| ≤ b and |v[j]| ≤ b, then a[i][j] − u[i] and a[i][j] − v[j] lie
// in [−(A+b), A+b], and a[i][j] − u[i] − v[j] in [−(A+2b), A+2b], where
// A + 2b ≤ maxT − 1. So none of them overflows, and none equals maxT, the
// value path gives pi before the first scan. These are the only differences
// of costs and dual prices that path, initialize, minCol and reducedCost
// compute. reducedCost skips removed edges, so its |a[i][j]| ≤ a0 holds even
// after a removal that assign hasn't seen yet.
//
// The dual prices change only in initialize, which checks each one against b
// before using it, and in path. After initialize, path only raises u and
// lowers v, by d = min pi[j] over the unlabelled columns. Each pi[j] is a
// reduced cost, so 0 ≤ d ≤ maxT − 1 by dual feasibility and the bound above.
// So path checks only d ≤ b − u[i] and d ≤ v[j] + b. With |u[i]|, |v[j]| ≤ b,
// both right-hand sides lie in [0, 2b] and can't overflow. path also lowers
// each unlabelled pi[j] by d, and pi[j] ≥ d, so pi[j] − d ≥ 0. remove and
// increase don't touch the dual prices, but remove can raise A, which lowers
// b, so assign recomputes b and rechecks every dual price before solving.
//
// a0 ≤ maxT, since no cost is the smallest value of T, and |m| ≤ maxT − 1,
// since minC < m < maxT. So maxT − A − 1 doesn't overflow, and it is
// negative only if A = maxT.
func (a *dense[T]) margin() T {
	bigA := a.a0
	if a.removed {
		bigA = max(bigA, abs(a.m))
	}
	if bigA == a.maxT {
		return -1
	}
	return (a.maxT - bigA - 1) / 2
}

// inMargin reports whether |x| ≤ a.b, without computing |x|.
func (a *dense[T]) inMargin(x T) bool {
	return -a.b <= x && x <= a.b
}

// bigM returns n·(maxC − minC) + maxC + 1, and false if that overflows T.
// Any assignment that uses an edge of cost M costs more than any assignment
// that doesn't, since M + (n−1)·minC > n·maxC.
//
// It adds maxC − minC ≥ 0 to maxC + 1 n times, so the partial sums rise to M,
// and one overflows only if M does. Multiplying first would overflow for some
// M that fit when maxC < 0. If maxC − minC overflows, then maxC ≥ 0, so
// M ≥ (maxC − minC) + maxC + 1 overflows too.
func bigM[T ap.Integer](n int, minC, maxC T) (T, bool) {
	r, ok := intmath.Sub(maxC, minC)
	if !ok {
		return 0, false
	}
	m, ok := intmath.Add(maxC, 1)
	for k := 0; ok && k < n; k++ {
		m, ok = intmath.Add(m, r)
	}
	return m, ok
}

// assign rechecks the dual prices against the current margin, solves, and
// sums the cost of the assignment with checked additions.
func (a *dense[T]) assign() (ap.Permutation, error) {
	if !a.overflow {
		a.b = a.margin()
		for i := 0; i < a.n; i++ {
			if !a.inMargin(a.u[i]) || !a.inMargin(a.v[i]) {
				a.overflow = true
				break
			}
		}
	}
	if !a.overflow {
		a.solve()
	}
	if a.overflow {
		return nil, ap.ErrOverflow
	}

	for i, j := range a.f {
		if a.a[i][j] == a.m {
			return nil, ap.ErrInfeasible
		}
	}

	var z T
	for i, j := range a.f {
		var ok bool
		if z, ok = intmath.Add(z, a.a[i][j]); !ok {
			a.overflow = true
			return nil, ap.ErrOverflow
		}
	}
	a.z = z

	p := make(ap.Permutation, a.n)
	copy(p, a.f)
	return p, nil
}

func (a *dense[T]) cost() T {
	return a.z
}

func (a *dense[T]) dualPrices() ap.DualPrices[T] {
	u := make([]T, a.n)
	v := make([]T, a.n)
	copy(u, a.u)
	copy(v, a.v)
	return ap.DualPrices[T]{U: u, V: v}
}

func (a *dense[T]) inRange(i, j int) bool {
	return i >= 0 && i < a.n && j >= 0 && j < a.n
}

func (a *dense[T]) reducedCost(i, j int) (T, bool) {
	if !a.inRange(i, j) || a.a[i][j] == a.m {
		return 0, false
	}
	return a.a[i][j] - a.u[i] - a.v[j], true
}

// remove sets the cost of edge (i, j) to m, first copying row i if it is
// shared with another kernel. If the edge is assigned, it is unassigned, so
// the next solve re-optimizes row i.
func (a *dense[T]) remove(i, j int) {
	if !a.inRange(i, j) || a.a[i][j] == a.m {
		return
	}
	if !a.owned[i].Load() {
		a.a[i] = append([]T(nil), a.a[i]...)
		a.owned[i].Store(true)
	}
	if a.f[i] == j {
		a.f[i] = -1
		a.fBar[j] = -1
	}
	a.a[i][j] = a.m
	a.removed = true
}

// copy shares the cost rows between a and the copy, and marks them as not
// owned in both, so that whichever kernel next removes an edge from a row
// copies that row first. Apart from the atomic stores to a.owned, it only
// reads a, so concurrent calls to copy on the same kernel are safe.
func (a *dense[T]) copy() kernel[T] {
	for i := range a.owned {
		a.owned[i].Store(false)
	}

	return &dense[T]{
		m:    a.m,
		maxT: a.maxT,
		n:    a.n,

		a0:       a.a0,
		b:        a.b,
		removed:  a.removed,
		overflow: a.overflow,

		a:     append([][]T(nil), a.a...),
		owned: make([]atomic.Bool, a.n),
		u:     append([]T(nil), a.u...),
		v:     append([]T(nil), a.v...),
		z:     a.z,

		f:    append([]int(nil), a.f...),
		fBar: append([]int(nil), a.fBar...),
		p:    append([]int(nil), a.p...),
		c:    append([]int(nil), a.c...),
		pi:   append([]T(nil), a.pi...),
	}
}
