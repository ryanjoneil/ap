package lsap

// Copy duplicates an assignment problem.
func (a *LSAP) Copy() *LSAP {
	A := make([][]int64, a.n)
	u := make([]int64, a.n)
	v := make([]int64, a.n)
	f := make([]int, a.n)
	fBar := make([]int, a.n)
	p := make([]int, a.n)
	c := make([]int, a.n)
	pi := make([]int64, a.n)

	for i := 0; i < a.n; i++ {
		A[i] = make([]int64, a.n)
		copy(A[i], a.a[i])
	}

	copy(u, a.u)
	copy(v, a.v)
	copy(f, a.f)
	copy(fBar, a.fBar)
	copy(p, a.p)
	copy(c, a.c)
	copy(pi, a.pi)

	return &LSAP{
		M: a.M,

		a: A,
		u: u,
		v: v,
		z: a.z,

		f:    f,
		fBar: fBar,
		p:    p,
		c:    c,
		pi:   pi,

		n: a.n,
	}
}
