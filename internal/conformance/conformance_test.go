package conformance

import (
	"github.com/ryanjoneil/ap"
	"github.com/ryanjoneil/ap/lsap"
	"github.com/ryanjoneil/ap/matching"
)

type solver[S any] interface {
	Assign() (ap.Permutation, error)
	Remove(u, v int)
	Copy() S
}

type coster[T any] interface{ Cost() T }

var (
	_ solver[*lsap.LSAP[int64]] = (*lsap.LSAP[int64])(nil)
	_ coster[int64]             = (*lsap.LSAP[int64])(nil)
	_ solver[*matching.Matcher] = (*matching.Matcher)(nil)
)
