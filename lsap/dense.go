package lsap

import (
	"fmt"

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
// The cost rows are copied on write, so that Copy is cheap.
type dense[T ap.Integer] struct {
	m    T   // Cost of a removed edge; every real cost is below m
	maxT T   // Largest value of T, cached for path
	n    int // n of assignment problem

	a     [][]T  // a[i][j] = cost of assigning row i to column j
	owned []bool // owned[i] = true if a[i] is not shared with another kernel
	u     []T    // u[i] = dual price for row i
	v     []T    // v[j] = dual price for column j
	z     T      // Objective value

	f    []int // f[i] = column assigned to row i, -1 if unassigned
	fBar []int // fBar[j] = row assigned to column j, -1 if unassigned
	p    []int // First column of row i not yet examined in phase 2
	c    []int // c[j] = row preceding column j in current alternating path
	pi   []T   // pi[j] = min { a[i,j] - u[i] - v[j] | i in lr, i != fBar[j] }
}

// newDense copies the square, non-empty cost matrix costs into a dense kernel
// and initializes it. It returns an error if costs could overflow T while
// solving; see New for the rule.
func newDense[T ap.Integer](costs [][]T) (*dense[T], error) {
	n := len(costs)

	minC, maxC := costs[0][0], costs[0][0]
	for _, row := range costs {
		for _, cij := range row {
			minC = min(minC, cij)
			maxC = max(maxC, cij)
		}
	}

	m, ok := bigM(n, minC, maxC)
	k := 4 * (n + 1)
	if !ok || !intmath.FitsScaled(k, minC) || !intmath.FitsScaled(k, m) {
		return nil, fmt.Errorf(
			"lsap: cost range [%d, %d] could overflow %T for n = %d",
			minC, maxC, minC, n,
		)
	}

	a := make([][]T, n)
	owned := make([]bool, n)
	for i, row := range costs {
		a[i] = append([]T(nil), row...)
		owned[i] = true
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
		maxT:  intmath.MaxOf[T](),
		n:     n,
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
	d.initialize()
	return d, nil
}

// bigM returns n·(maxC − minC) + maxC + 1, and false if that overflows T.
// Any assignment that uses an edge of cost M costs more than any assignment
// that doesn't, since M + (n−1)·minC > n·maxC.
func bigM[T ap.Integer](n int, minC, maxC T) (T, bool) {
	r, ok := intmath.Sub(maxC, minC)
	if !ok {
		return 0, false
	}
	if r, ok = intmath.MulInt(r, n); !ok {
		return 0, false
	}
	if r, ok = intmath.Add(r, maxC); !ok {
		return 0, false
	}
	return intmath.Add(r, 1)
}

func (a *dense[T]) assign() (ap.Permutation, error) {
	a.solve()
	for i, j := range a.f {
		if a.a[i][j] == a.m {
			return nil, ap.ErrInfeasible
		}
	}
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
	if !a.owned[i] {
		a.a[i] = append([]T(nil), a.a[i]...)
		a.owned[i] = true
	}
	if a.f[i] == j {
		a.f[i] = -1
		a.fBar[j] = -1
	}
	a.a[i][j] = a.m
}

// copy shares the cost rows between a and the copy, and marks them as not
// owned in both, so that whichever kernel next removes an edge from a row
// copies that row first.
func (a *dense[T]) copy() kernel[T] {
	for i := range a.owned {
		a.owned[i] = false
	}

	return &dense[T]{
		m:    a.m,
		maxT: a.maxT,
		n:    a.n,

		a:     append([][]T(nil), a.a...),
		owned: make([]bool, a.n),
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
