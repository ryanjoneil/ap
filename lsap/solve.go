package lsap

// solve updates f, fBar, u and v to the optimal LSAP solution. It is the main
// loop of Hungarian_3 (book §4.2.2, Algorithm 4.5) as coded in APC by
// Carpaneto, Martello, and Toth (1988). It stops early if path sets
// a.overflow. As a departure from APC, it doesn't compute the objective value;
// assign does that without overflow.
func (a *dense[T]) solve() {
	for i := 0; i < a.n; i++ {
		if a.f[i] < 0 {
			j := a.path(i)
			if j < 0 {
				return
			}
			a.increase(i, j)
		}
	}
}

// path grows an alternating tree from the unassigned row i, updating the dual
// prices, until it reaches an unassigned column, which it returns. It is
// Procedure Augment(k) (book §4.2.2, Algorithm 4.4). Where the book picks any
// labelled, unscanned column, path picks the one with the lowest index.
//
// As a departure from the source, path checks that each dual update keeps the
// dual prices within the margin a.b (see margin), which makes the unchecked
// reduced costs in its inner loop safe. If an update would leave the margin,
// path sets a.overflow and returns −1 without making it.
//
// As a departure from the source, path removes every required column from uc
// before the search starts, so it never labels one, and never reaches a
// required row, whose only way in is its required column (see dense). The
// dual update skips required columns, so it leaves their v[j] and pi[j]
// alone.
func (a *dense[T]) path(i int) int {
	lr := []int{i}          // Vector of labelled rows
	uc := make([]bool, a.n) // uc[j] = true if column j is unlabelled
	fixed := a.fixRow       // fixed[j] ≥ 0 if column j is required; nil if none is

	for j := 0; j < a.n; j++ {
		uc[j] = true
		a.pi[j] = a.maxT
	}
	for j, r := range fixed {
		if r >= 0 {
			uc[j] = false
		}
	}

	for {
		r := lr[len(lr)-1]

		// Locals: otherwise the loop reloads each through a, and its speed
		// is layout-sensitive.
		ar, ur, v, pi, c := a.a[r], a.u[r], a.v, a.pi, a.c
		for j := range uc {
			if !uc[j] {
				continue
			}
			val := ar[j] - ur - v[j]
			if val < pi[j] {
				pi[j] = val
				c[j] = r
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

			// d ≥ 0, so u only rises and v only falls. Neither check can
			// overflow, since b − u[h] and v[j] + b lie in [0, 2b].
			for _, h := range lr {
				if d > a.b-a.u[h] {
					a.overflow = true
					return -1
				}
				a.u[h] += d
			}

			// The columns with pi[j] == 0 are the labelled ones, apart from
			// the required ones, which are skipped. Every unlabelled column
			// has pi[j] ≥ d, as d is their minimum, so pi[j] − d ≥ 0.
			for j := 0; j < a.n; j++ {
				if fixed != nil && fixed[j] >= 0 {
					continue
				}
				if a.pi[j] == 0 {
					if d > a.v[j]+a.b {
						a.overflow = true
						return -1
					}
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
