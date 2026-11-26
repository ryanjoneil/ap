package intmath

import (
	"math"
	"testing"
)

func TestBounds(t *testing.T) {
	if MaxOf[int8]() != math.MaxInt8 || MinOf[int8]() != math.MinInt8 {
		t.Error("int8")
	}
	if MaxOf[int16]() != math.MaxInt16 || MinOf[int16]() != math.MinInt16 {
		t.Error("int16")
	}
	if MaxOf[int32]() != math.MaxInt32 || MinOf[int32]() != math.MinInt32 {
		t.Error("int32")
	}
	if MaxOf[int64]() != math.MaxInt64 || MinOf[int64]() != math.MinInt64 {
		t.Error("int64")
	}
	if MaxOf[int]() != math.MaxInt || MinOf[int]() != math.MinInt {
		t.Error("int")
	}
}

func TestChecked(t *testing.T) {
	if _, ok := Add[int8](100, 27); !ok {
		t.Error("100+27 fits")
	}
	if _, ok := Add[int8](100, 28); ok {
		t.Error("100+28 overflows")
	}
	if _, ok := Sub[int8](-100, 28); !ok {
		t.Error("-100-28 fits")
	}
	if _, ok := Sub[int8](-100, 29); ok {
		t.Error("-100-29 overflows")
	}
	if v, ok := MulInt[int8](-42, 3); !ok || v != -126 {
		t.Error("-42*3")
	}
	if _, ok := MulInt[int8](43, 3); ok {
		t.Error("43*3 overflows")
	}
}

func TestFitsScaled(t *testing.T) {
	cases := []struct {
		k    int
		c    int8
		want bool
	}{
		{1, 127, true}, {2, 63, true}, {2, 64, false}, {2, -63, true},
		{1, -128, false}, {0, 1, false}, {-1, 1, false}, {1000, 0, true},
	}
	for _, c := range cases {
		if got := FitsScaled(c.k, c.c); got != c.want {
			t.Errorf("FitsScaled(%d, %d) = %v", c.k, c.c, got)
		}
	}
}
