package lsap

// Copy duplicates an assignment problem.
func (a *LSAP) Copy() *LSAP {
	A := make([][]int64, a.Size)
	U := make([]int64, a.Size)
	V := make([]int64, a.Size)
	f := make([]int, a.Size)
	fBar := make([]int, a.Size)
	p := make([]int, a.Size)
	c := make([]int, a.Size)
	pi := make([]int64, a.Size)

	for i := 0; i < a.Size; i++ {
		A[i] = make([]int64, a.Size)
		copy(A[i], a.A[i])
	}

	copy(U, a.U)
	copy(V, a.V)
	copy(f, a.f)
	copy(fBar, a.fBar)
	copy(p, a.p)
	copy(c, a.c)
	copy(pi, a.pi)

	return &LSAP{
		M: a.M,

		A: A,
		U: U,
		V: V,
		Z: a.Z,

		f:    f,
		fBar: fBar,
		p:    p,
		c:    c,
		pi:   pi,

		Size: a.Size,
	}
}
