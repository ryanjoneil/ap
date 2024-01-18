package ap

// DualPrices are dual prices for the assignment constraints corresponding to
// the U and V sets, respectively.
type DualPrices[T Integer] struct {
	U []T `json:"u"`
	V []T `json:"v"`
}

// A DualPricer provides dual prices on the assignment constraints associated
// with sets U and V. A dual price is the value of a unit of slack on a binding
// constraint.
type DualPricer[T Integer] interface {
	DualPrices() DualPrices[T]
}
