package lsap

// solve updates A, U, V, and Z to the optimal LSAP solution. It is the main
// loop of Hungarian_3 (book §4.2.2, Algorithm 4.5) as coded in APC by
// Carpaneto, Martello, and Toth (1988).
func (a *dense[T]) solve() {
	for i := 0; i < a.n; i++ {
		if a.f[i] < 0 {
			j := a.path(i)
			if j >= 0 {
				a.increase(i, j)
			}
		}
	}

	a.z = 0
	for i := 0; i < a.n; i++ {
		a.z += a.u[i] + a.v[i]
	}
}

// path grows an alternating tree from the unassigned row i, updating the dual
// prices, until it reaches an unassigned column, which it returns. It is
// Procedure Augment(k) (book §4.2.2, Algorithm 4.4). Where the book picks any
// labelled, unscanned column, path picks the one with the lowest index.
func (a *dense[T]) path(i int) int {
	lr := []int{i}          // Vector of labelled rows
	uc := make([]bool, a.n) // uc[j] = true if column j is unlabelled

	for j := 0; j < a.n; j++ {
		uc[j] = true
		a.pi[j] = a.maxT
	}

	for {
		r := lr[len(lr)-1]

		for j := 0; j < a.n; j++ {
			if !uc[j] {
				continue
			}
			val := a.a[r][j] - a.u[r] - a.v[j]
			if val < a.pi[j] {
				a.pi[j] = val
				a.c[j] = r
			}
		}

		found := false
		for j := 0; j < a.n; j++ {
			if uc[j] && a.pi[j] == 0 {
				found = true
				break
			}
		}

		if !found {
			// d = min { pi[j] | j in uc }
			first := true
			var d T
			for j := 0; j < a.n; j++ {
				if !uc[j] {
					continue
				}
				if first || a.pi[j] < d {
					first = false
					d = a.pi[j]
				}
			}

			for _, h := range lr {
				a.u[h] += d
			}

			for j := 0; j < a.n; j++ {
				if a.pi[j] == 0 {
					a.v[j] -= d
				} else {
					a.pi[j] -= d
				}
			}
		}

		// j = first column in { k in uc | pi[k] = 0 }
		j := -1
		for k := 0; k < a.n; k++ {
			if uc[k] && a.pi[k] == 0 {
				j = k
				break
			}
		}

		if j >= 0 && a.fBar[j] >= 0 {
			lr = append(lr, a.fBar[j])
			uc[j] = false
		}

		if j >= 0 && a.fBar[j] < 0 {
			return j
		}
	}
}

// increase augments the assignment along the path from row i to column j
// found by path (the repeat loop of book Algorithm 4.5).
func (a *dense[T]) increase(i, j int) {
	for {
		l := a.c[j]
		a.fBar[j] = l
		k := a.f[l]
		a.f[l] = j
		j = k

		if l == i {
			break
		}
	}
}
