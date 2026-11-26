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

func TestAddWrap(t *testing.T) {
	cases := []struct {
		a, b  int8
		sum   int8
		carry int
	}{
		{100, 27, 127, 0}, {100, 28, -128, 1}, {127, 127, -2, 1},
		{-100, -28, -128, 0}, {-100, -29, 127, -1}, {-128, -128, 0, -1},
		{-128, 127, -1, 0}, {5, 0, 5, 0}, {0, -128, -128, 0},
	}
	for _, c := range cases {
		if sum, carry := AddWrap(c.a, c.b); sum != c.sum || carry != c.carry {
			t.Errorf("AddWrap(%d, %d) = (%d, %d); want (%d, %d)", c.a, c.b, sum, carry, c.sum, c.carry)
		}
	}
}
