package ap

// DualPrices are dual prices for the assignment constraints corresponding to
// the U and V sets, respectively.
type DualPrices[T Integer] struct {
	U []T `json:"u"`
	V []T `json:"v"`
}
