package lsap

func (a *LSAP) initialize() {
	a.initializePhase1()
	a.initializePhase2()
}

func (a *LSAP) initializePhase1() {
	for j := 0; j < a.Size; j++ {
		r := a.minRow(j)
		a.V[j] = a.A[r][j]
		if a.f[r] < 0 {
			a.fBar[j] = r
			a.f[r] = j
			a.U[r] = 0
			a.p[r] = j + 1
		}
	}
}

func (a *LSAP) initializePhase2() {
	for i := 0; i < a.Size; i++ {
		if a.f[i] >= 0 {
			continue
		}

		// Determine value of U[i].
		j := a.minCol(i)
		a.U[i] = a.A[i][j] - a.V[j]

		assign := a.fBar[j] < 0
		for !assign && j < a.Size {
			if a.A[i][j]-a.U[i]-a.V[j] == 0 {
				r := a.fBar[j]
				if r < 0 {
					break
				}
				k := a.p[r]

				for !assign && k < a.Size {
					if r >= 0 && k >= 0 && a.fBar[k] < 0 && a.A[r][k]-a.U[r]-a.V[k] == 0 {
						assign = true
						a.f[r] = k
						a.fBar[k] = r
					} else {
						k++
					}
				}

				// Note: The paper sets p[r] = k+1, but the FORTRAN code
				//       sets p[r] = n. We use the latter as our standard.
				a.p[r] = a.Size
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

func (a *LSAP) minRow(j int) int {
	first := true
	var minVal int64
	minRow := 0

	for i := 0; i < a.Size; i++ {
		if first || a.A[i][j] < minVal || (a.A[i][j] == minVal && a.f[i] < 0) {
			first = false
			minVal = a.A[i][j]
			minRow = i
		}
	}

	return minRow
}

func (a *LSAP) minCol(i int) int {
	first := true
	var minVal int64
	minCol := 0

	for j := 0; j < a.Size; j++ {
		if first || a.A[i][j]-a.V[j] < minVal || a.A[i][j] == minVal && a.fBar[j] < 0 {
			first = false
			minVal = a.A[i][j] - a.V[j]
			minCol = j
		}
	}

	return minCol
}
