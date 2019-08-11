package ap

// An Assigner creates a mapping between two sets U = V = {0,...,n-1}.
type Assigner interface {
	Assign() Permutation
}
