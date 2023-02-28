package lsap

import (
	"math"

	"github.com/ryanjoneil/ap"
)

// An Assigner for linear sum assignment problem.
type Assigner interface {
	ap.Assigner
	ap.Int64Coster
	ap.Int64DualPricer
	ap.Int64ReducedCoster

	Copy() Assigner
	Remove(int, int)
}

// New linear sum assignment problem (LSAP) from a square cost matrix.
// Note: mutates the cost matrix.
func New(A [][]int64) Assigner {
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
	pi := make([]int64, n)

	for i := 0; i < n; i++ {
		f[i] = -1
		fBar[i] = -1
		p[i] = -1
		c[i] = -1
		pi[i] = -1
	}

	a := &lsap{
		M:    math.MaxInt64 / int64(n),
		a:    A,
		u:    make([]int64, n),
		v:    make([]int64, n),
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
