package lsap

// initialize finds a partial assignment and feasible dual prices that satisfy
// complementary slackness. It is three edge preprocessing (book §4.4.4,
// Algorithm 4.11), as coded in APC by Carpaneto, Martello, and Toth (1988).
//
// As a departure from the source, it checks each dual price against the
// margin a.b (see margin) as it sets it, before the price is used. If one is
// outside the margin, initialize sets a.overflow and stops.
func (a *dense[T]) initialize() {
	if a.initializePhase1() {
		a.initializePhase2()
	}
}

// initializePhase1 sets each v[j] to the smallest cost in column j, and
// assigns that row to column j if the row is still unassigned. It returns
// false if some v[j] is outside the margin.
func (a *dense[T]) initializePhase1() bool {
	for j := 0; j < a.n; j++ {
		r := a.minRow(j)
		if !a.inMargin(a.a[r][j]) {
			a.overflow = true
			return false
		}
		a.v[j] = a.a[r][j]
		if a.f[r] < 0 {
			a.fBar[j] = r
			a.f[r] = j
			a.u[r] = 0
			a.p[r] = j + 1
		}
	}
	return true
}

// initializePhase2 sets u[i] for each unassigned row i, and tries to assign i,
// reassigning one row if needed. It stops if some u[i] is outside the margin.
func (a *dense[T]) initializePhase2() {
	for i := 0; i < a.n; i++ {
		if a.f[i] >= 0 {
			continue
		}

		// Determine value of U[i].
		j := a.minCol(i)
		ui := a.a[i][j] - a.v[j]
		if !a.inMargin(ui) {
			a.overflow = true
			return
		}
		a.u[i] = ui

		assign := a.fBar[j] < 0
		for !assign && j < a.n {
			if a.a[i][j]-a.u[i]-a.v[j] == 0 {
				r := a.fBar[j]
				if r < 0 {
					break
				}
				k := a.p[r]

				for !assign && k < a.n {
					if r >= 0 &&
						k >= 0 &&
						a.fBar[k] < 0 &&
						a.a[r][k]-a.u[r]-a.v[k] == 0 {

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

func (a *dense[T]) minRow(j int) int {
	first := true
	var minVal T
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

func (a *dense[T]) minCol(i int) int {
	first := true
	var minVal T
	minCol := 0

	for j := 0; j < a.n; j++ {
		if first ||
			a.a[i][j]-a.v[j] < minVal ||
			a.a[i][j]-a.v[j] == minVal && a.fBar[j] < 0 {

			first = false
			minVal = a.a[i][j] - a.v[j]
			minCol = j
		}
	}

	return minCol
}
