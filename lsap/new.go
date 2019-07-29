package lsap

import "math"

// New linear sum assignment problem (LSAP) from a square cost matrix.
// Note: mutates the cost matrix.
func New(A [][]int64) *LSAP {
	size := len(A)
	if size < 1 {
		panic("empty cost matrix")
	}
	for _, row := range A {
		if len(row) != size {
			panic("cost matrix not square")
		}
	}

	f := make([]int, size)
	fBar := make([]int, size)
	p := make([]int, size)
	c := make([]int, size)
	pi := make([]int64, size)

	for i := 0; i < size; i++ {
		f[i] = -1
		fBar[i] = -1
		p[i] = -1
		c[i] = -1
		pi[i] = -1
	}

	a := &LSAP{
		M:    int64(math.Pow(1000, 3)),
		A:    A,
		U:    make([]int64, size),
		V:    make([]int64, size),
		f:    f,
		fBar: fBar,
		p:    p,
		c:    c,
		pi:   pi,
		Size: size,
	}
	a.initialize()
	return a
}
