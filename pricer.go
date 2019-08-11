package ap

// Int64DualPrices are dual prices for the assignment constraints corresponding
// to the U and V sets, respectively.
type Int64DualPrices struct {
	U []int64
	V []int64
}

// An Int64DualPricer provides dual prices on the assignment constraints
// associated with sets U and V. A dual price is the value of a unit of slack on
// a binding constraint.
type Int64DualPricer interface {
	DualPrices() Int64DualPrices
}
