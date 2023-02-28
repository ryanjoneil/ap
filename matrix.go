package ap

// Arc of a matrix connects row to column and is true if the value is nonzero.
type Arc int8

func (a Arc) Bool() bool {
	return a != 0
}

// Matrix representation of an assignment. If u is assigned to v, then M[u][v]
// is true. Each row and each column has exactly one true element.
type Matrix [][]Arc

// Cycles converts a permutation matrix to a cyclic representation.
func (m Matrix) Cycles() Cycles {
	return m.Permutation().Cycles()
}

// Inverse inverts a permutation matrix, changing its direction.
func (m Matrix) Inverse() Matrix {
	return m.Permutation().Inverse().Matrix()
}

// Permutation converts a matrix assignment representation into a permutation.
func (m Matrix) Permutation() Permutation {
	p := make(Permutation, len(m))
	for u, mu := range m {
		for v, assigned := range mu {
			if assigned != 0 {
				p[u] = v
				break
			}
		}
	}
	return p
}
