package ap

import (
	"math"
)

// Solve updates A, U, V, and Z to the optimal AP solution.
func (ap *AP) Solve() {
	if !ap.initialized {
		ap.initialize()
	}

	for i := 0; i < ap.Size; i++ {
		if ap.f[i] < 0 {
			j := ap.path(i)
			if j >= 0 {
				ap.increase(i, j)
			}
		}
	}

	ap.Z = 0
	for i := 0; i < ap.Size; i++ {
		ap.Z += ap.U[i] + ap.V[i]
	}
}

func (ap *AP) path(i int) int {
	ap.lr = []int{i}
	for j := 0; j < ap.Size; j++ {
		ap.uc[j] = true
		ap.pi[j] = math.MaxInt64
	}

	for {
		r := ap.lr[len(ap.lr)-1]
		if r >= ap.Size {
			break
		}

		for j := range ap.uc {
			val := ap.A[r][j] - ap.U[r] - ap.V[j]
			if val < ap.pi[j] {
				ap.pi[j] = val
				ap.c[j] = r
			}
		}

		found := false
		for j := range ap.uc {
			if ap.pi[j] == 0 {
				found = true
				break
			}
		}

		if !found {
			// d = min { pi[j] | j in uc }
			first := true
			d := 0
			for j := range ap.uc {
				if first || ap.pi[j] < d {
					first = false
					d = ap.pi[j]
				}
			}

			for _, h := range ap.lr {
				ap.U[h] += d
			}

			for j := 0; j < ap.Size; j++ {
				if ap.pi[j] == 0 {
					ap.V[j] -= d
				} else {
					ap.pi[j] -= d
				}
			}
		}

		// j = first column in { k in uc | pi[k] = 0 }
		j := -1
		for k := range ap.uc {
			if ap.pi[k] == 0 {
				j = k
				break
			}
		}

		if j >= 0 && ap.fBar[j] >= 0 {
			ap.lr = append(ap.lr, ap.fBar[j])
			delete(ap.uc, j)
		}

		if j >= 0 && ap.fBar[j] < 0 {
			return j
		}
	}

	return -1
}

func (ap *AP) increase(i, j int) {
	for {
		l := ap.c[j]
		ap.fBar[j] = l
		k := ap.f[l]
		ap.f[l] = j
		j = k

		if l == i {
			break
		}
	}
}
