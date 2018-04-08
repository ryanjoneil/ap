package ap

// Copy makes an exact duplicate of an AP.
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
		for j := 0; j < ap.Size; j++ {
			A[i][j] = ap.A[i][j]
		}

		U[i] = ap.U[i]
		V[i] = ap.V[i]
		f[i] = ap.f[i]
		fBar[i] = ap.fBar[i]
		p[i] = ap.p[i]
		c[i] = ap.c[i]
		pi[i] = ap.pi[i]
	}

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
