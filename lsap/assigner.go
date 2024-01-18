package lsap

import (
	"github.com/ryanjoneil/ap"
)

// An Assigner for linear sum assignment problem conforms to the [ap.Coster],
// [ap.DualPricer] and [ap.ReducedCoster] interfaces. Dual bounds and reduced
// costs are calculated as part of the assignment optimization. A copy method
// allows one to integrate an LSAP solver into a search routine for reduced
// cost-based domain filtering.
type Assigner[T ap.Integer] interface {
	ap.Assigner
	ap.Coster[T]
	ap.DualPricer[T]
	ap.ReducedCoster[T]

	Copy() Assigner[T]
	M() T
	Remove(int, int)
	SetM(T)
}

// New linear sum assignment problem (LSAP) from a square cost matrix.
// Note: mutates the cost matrix.
func New[T ap.Integer](A [][]T) Assigner[T] {
	n := len(A)
	if n < 1 {
		panic("empty cost matrix")
	}

	for _, row := range A {
		if len(row) != n {
			panic("cost matrix not square")
		}
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

	a := &LSAP[T]{
		m:    ap.MaxOf[T]() / T(n),
		a:    A,
		u:    make([]T, n),
		v:    make([]T, n),
		f:    f,
		fBar: fBar,
		p:    p,
		c:    c,
		pi:   pi,
		n:    n,
	}
	a.initialize()
	return a
}
