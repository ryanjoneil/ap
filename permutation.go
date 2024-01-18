package ap

// A Permutation P maps elements from one set U = {0,...,n-1} to another set
// V = {0,...,n-1}, where P[u] = v means that u ∈ U is assigned to v ∈ V.
//
//	p := ap.Permutation{1, 0, 2} // Assign 0 to 1, 1 to 0, and 2 to itself.
type Permutation []int

// Cycles convert a permutation representation of an assignment to a cyclical
// representation of that assignment. If the result contains a single cycle,
// then that cycle is a Hamiltonian cycle.
//
//	p := ap.Permutation{1, 0, 2, 6, 5, 3, 4}
//	p.Cycles() // {{0, 1}, {2}, {3, 6, 4, 5}}
func (p Permutation) Cycles() Cycles {
	cycles := Cycles{}
	seen := make([]bool, len(p))

	c := Cycle{}
	for start := range p {
		for i := start; !seen[i]; i = p[i] {
			seen[i] = true
			c = append(c, i)
		}
		if len(c) > 0 {
			cycles = append(cycles, c)
			c = Cycle{}
		}
	}

	return cycles
}

// Inverse converts an assignment from a set U to a set V to an assignment from
// V to U. If, for example, U is the left hand side of a bipartite matching and
// V is the right hand side, this function essentially swaps their sides.
func (p Permutation) Inverse() Permutation {
	p2 := make(Permutation, len(p))
	for u, v := range p {
		p2[v] = u
	}
	return p2
}

// Matrix converts a permutation into a square matrix.
func (p Permutation) Matrix() Matrix {
	m := make(Matrix, len(p))
	for u, v := range p {
		m[u] = make([]bool, len(p))
		m[u][v] = true
	}
	return m
}
