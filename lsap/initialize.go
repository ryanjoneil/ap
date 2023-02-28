package lsap

func (a *lsap) initialize() {
	a.initializePhase1()
	a.initializePhase2()
}

func (a *lsap) initializePhase1() {
	for j := 0; j < a.n; j++ {
		r := a.minRow(j)
		a.v[j] = a.a[r][j]
		if a.f[r] < 0 {
			a.fBar[j] = r
			a.f[r] = j
			a.u[r] = 0
			a.p[r] = j + 1
		}
	}
}

func (a *lsap) initializePhase2() {
	for i := 0; i < a.n; i++ {
		if a.f[i] >= 0 {
			continue
		}

		// Determine value of U[i].
		j := a.minCol(i)
		a.u[i] = a.a[i][j] - a.v[j]

		assign := a.fBar[j] < 0
		for !assign && j < a.n {
			if a.a[i][j]-a.u[i]-a.v[j] == 0 {
				r := a.fBar[j]
				if r < 0 {
					break
				}
				k := a.p[r]

				for !assign && k < a.n {
					if r >= 0 && k >= 0 && a.fBar[k] < 0 && a.a[r][k]-a.u[r]-a.v[k] == 0 {
						assign = true
						a.f[r] = k
						a.fBar[k] = r
					} else {
						k++
					}
				}

				// Note: The paper sets p[r] = k+1, but the FORTRAN code
				//       sets p[r] = n. We use the latter as our standard.
				a.p[r] = a.n
			}

			if !assign {
				j++
			}
		}

		if assign {
			a.f[i] = j
			a.fBar[j] = i
			a.p[i] = j + 1
		}
	}
}

func (a *lsap) minRow(j int) int {
	first := true
	var minVal int64
	minRow := 0

	for i := 0; i < a.n; i++ {
		if first || a.a[i][j] < minVal || (a.a[i][j] == minVal && a.f[i] < 0) {
			first = false
			minVal = a.a[i][j]
			minRow = i
		}
	}

	return minRow
}

func (a *lsap) minCol(i int) int {
	first := true
	var minVal int64
	minCol := 0

	for j := 0; j < a.n; j++ {
		if first || a.a[i][j]-a.v[j] < minVal || a.a[i][j] == minVal && a.fBar[j] < 0 {
			first = false
			minVal = a.a[i][j] - a.v[j]
			minCol = j
		}
	}

	return minCol
}
