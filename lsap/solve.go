package lsap

import "github.com/ryanjoneil/ap"

// solve updates A, U, V, and Z to the optimal LSAP solution.
func (a *LSAP[T]) solve() {
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

func (a *LSAP[T]) path(i int) int {
	lr := []int{i}          // Vector of labelled rows
	uc := make([]bool, a.n) // uc[j] = true if column j is unlabelled

	for j := 0; j < a.n; j++ {
		uc[j] = true
		a.pi[j] = ap.MaxOf[T]()
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

func (a *LSAP[T]) increase(i, j int) {
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
