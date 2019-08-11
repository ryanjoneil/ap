package ap

// A Cycle is an individual cycle in an assignment, such as {1,2,4}. The cycle
// is assumed to contain an arc from the final element to the initial element,
// so this cycle has the arcs (1,2), (2,4) and (4,1). A cycle can contain a
// a single element with an arc to itself, such as {3}.
type Cycle []int

func (c Cycle) inverse() Cycle {
	c2 := make(Cycle, 0, len(c))

	// Start at the same node.
	if len(c) > 0 {
		c2 = append(c2, c[0])
	}

	// Procede through the remaining nodes in reverse order.
	for i := len(c) - 1; i > 0; i-- {
		c2 = append(c2, c[i])
	}

	return c2
}

// Cycles represents multiple individual cycle structures corresponding to an
// assignment. For example, the permutation {1,0,2,6,5,3,4} is equivalent to the
// set of cycles {{0,1}, {2}, {3,6,4,5}}. If a permutation corresponds to a
// single cycle, then that cycle is a Hamiltonian cycle.
type Cycles []Cycle

// Inverse changes the direction of a set of cycles representing an assignment.
//
//     c := ap.Cycles{{0, 1}, {2}, {3, 6, 4, 5}}
//     c.Inverse() // {{0, 1}, {2}, {3, 5, 4, 6}}
func (c Cycles) Inverse() Cycles {
	c2 := make(Cycles, len(c))
	for i, cycle := range c {
		c2[i] = cycle.inverse()
	}
	return c2
}

// Matrix converts a cyclic representation of an assignment into a permutation
// matrix.
func (c Cycles) Matrix() Matrix {
	m := make(Matrix, c.len())
	for u := range m {
		m[u] = make([]bool, len(m))
	}

	for _, cycle := range c {
		for i, u := range cycle {
			v := cycle[0]
			if i < len(cycle)-1 {
				v = cycle[i+1]
			}
			m[u][v] = true
		}
	}

	return m
}

// Permutation converts a cyclic representation of an assignment into a
// permutation representation.
//
//     c := ap.Cycles{{0, 1}, {2}, {3, 6, 4, 5}}
//     c.Permutation() // {1, 0, 2, 6, 5, 3, 4}
func (c Cycles) Permutation() Permutation {
	p := make(Permutation, c.len())
	for _, cycle := range c {
		for i, u := range cycle {
			v := cycle[0]
			if i < len(cycle)-1 {
				v = cycle[i+1]
			}
			p[u] = v
		}
	}
	return p
}

func (c Cycles) len() int {
	l := 0
	for _, cycle := range c {
		l += len(cycle)
	}
	return l
}
