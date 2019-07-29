package ap

// Inverse converts an assignment from a set U to a set V to an assignment from
// V to U. If, for example, U is the left hand side of a bipartite matching and
// V is the right hand side, this function essentially swaps their sides.
//
//     a := ap.Assignment{1, 3, 2, 0}
//     ap.Inverse(a) // {3, 0, 2, 1}
func Inverse(a Assignment) Assignment {
	a2 := make(Assignment, len(a))
	for u, v := range a {
		a2[v] = u
	}
	return a2
}
