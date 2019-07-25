package ap

// New assignment problem from a square cost matrix.
func New(A [][]int64) AP {
	if len(A) < 1 {
		panic("empty cost matrix")
	}
	for _, row := range A {
		if len(row) != len(A) {
			panic("cost matrix not square")
		}
	}

	size := len(A)
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

	return AP{
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
}
