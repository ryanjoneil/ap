package ap

// Create instantiates an Assignment Problem of a given size.
func Create(size int) *AP {
	A := make([][]int64, size) // 0
	for r := 0; r < size; r++ {
		A[r] = make([]int64, size)
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

	return &AP{
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
