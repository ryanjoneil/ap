package ap

import "strings"

// Matrix representation of an assignment. If u is assigned to v, then M[u][v]
// is true. Each row and column has exactly one true element.
type Matrix [][]bool

// ToMatrix converts an assignment into a square matrix.
func ToMatrix(a Assignment) Matrix {
	m := make(Matrix, len(a))
	for u, v := range a {
		m[u] = make([]bool, len(a))
		m[u][v] = true
	}
	return m
}

func (m Matrix) String() string {
	s := make([]string, len(m))
	for u, mu := range m {
		s[u] = rowString(mu)
	}
	return strings.Join(s, "\n")
}

func rowString(row []bool) string {
	s := make([]string, len(row))
	for i, v := range row {
		s[i] = "-"
		if v {
			s[i] = "X"
		}
	}
	return strings.Join(s, " ")
}
