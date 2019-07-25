package ap

// Copy duplicates an assignment problem.
func (ap *AP) Copy() *AP {
	A := make([][]int64, ap.Size)
	U := make([]int64, ap.Size)
	V := make([]int64, ap.Size)
	f := make([]int, ap.Size)
	fBar := make([]int, ap.Size)
	p := make([]int, ap.Size)
	c := make([]int, ap.Size)
	pi := make([]int64, ap.Size)

	for i := 0; i < ap.Size; i++ {
		A[i] = make([]int64, ap.Size)
		copy(A[i], ap.A[i])
	}

	copy(U, ap.U)
	copy(V, ap.V)
	copy(f, ap.f)
	copy(fBar, ap.fBar)
	copy(p, ap.p)
	copy(c, ap.c)
	copy(pi, ap.pi)

	return &AP{
		A: A,
		U: U,
		V: V,
		Z: ap.Z,

		f:    f,
		fBar: fBar,
		p:    p,
		c:    c,
		pi:   pi,

		Size:        ap.Size,
		initialized: ap.initialized,
	}
}
