package ap

func (ap *AP) initialize() {
	ap.initializePhase1()
	ap.initializePhase2()

	ap.initialized = true
}

func (ap *AP) initializePhase1() {
	for j := 0; j < ap.Size; j++ {
		r := ap.minRow(j)
		ap.V[j] = ap.A[r][j]
		if ap.f[r] < 0 {
			ap.fBar[j] = r
			ap.f[r] = j
			ap.U[r] = 0
			ap.p[r] = j + 1
		}
	}
}

func (ap *AP) initializePhase2() {
	for i := 0; i < ap.Size; i++ {
		if ap.f[i] >= 0 {
			continue
		}

		// Determine value of U[i].
		j := ap.minCol(i)
		ap.U[i] = ap.A[i][j] - ap.V[j]

		assign := ap.fBar[j] < 0
		for !assign && j < ap.Size {
			if ap.A[i][j]-ap.U[i]-ap.V[j] == 0 {
				r := ap.fBar[j]
				if r < 0 {
					break
				}
				k := ap.p[r]

				for !assign && k < ap.Size {
					if r >= 0 && k >= 0 && ap.fBar[k] < 0 && ap.A[r][k]-ap.U[r]-ap.V[k] == 0 {
						assign = true
						ap.f[r] = k
						ap.fBar[k] = r
					} else {
						k++
					}
				}

				// Note: The paper sets p[r] = k+1, but the FORTRAN code
				//       sets p[r] = n. We use the latter as our standard.
				ap.p[r] = ap.Size
			}

			if !assign {
				j++
			}
		}

		if assign {
			ap.f[i] = j
			ap.fBar[j] = i
			ap.p[i] = j + 1
		}
	}
}

func (ap *AP) minRow(j int) int {
	first := true
	var minVal int64
	minRow := 0

	for i := 0; i < ap.Size; i++ {
		if first || ap.A[i][j] < minVal || (ap.A[i][j] == minVal && ap.f[i] < 0) {
			first = false
			minVal = ap.A[i][j]
			minRow = i
		}
	}

	return minRow
}

func (ap *AP) minCol(i int) int {
	first := true
	var minVal int64
	minCol := 0

	for j := 0; j < ap.Size; j++ {
		if first || ap.A[i][j]-ap.V[j] < minVal || ap.A[i][j] == minVal && ap.fBar[j] < 0 {
			first = false
			minVal = ap.A[i][j] - ap.V[j]
			minCol = j
		}
	}

	return minCol
}
