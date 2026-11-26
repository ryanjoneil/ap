package ap_test

import (
	"testing"

	"github.com/ryanjoneil/ap"
)

func sum[T ap.Integer](xs ...T) T {
	var s T
	for _, x := range xs {
		s += x
	}
	return s
}

type cents int64

func TestInteger(t *testing.T) {
	if sum[int](1, 2) != 3 || sum[int8](1, 2) != 3 || sum[int16](1, 2) != 3 ||
		sum[int32](1, 2) != 3 || sum[int64](1, 2) != 3 || sum[cents](1, 2) != 3 {
		t.Fail()
	}
}
