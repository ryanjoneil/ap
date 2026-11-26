// Package intmath provides bounds and overflow-checked arithmetic for the
// signed integer types accepted by the ap package.
package intmath

import (
	"math/bits"

	"github.com/ryanjoneil/ap"
)

// MaxOf returns the maximum value of the signed integer type T.
func MaxOf[T ap.Integer]() T {
	// 1 shifted into the sign bit is the minimum value, and its complement is
	// the maximum. 1 << w is negative only if w is the sign bit, so try the
	// sign bits of the possible widths in turn: 7, 15, 31, 63. The loop runs
	// at most four times.
	var one T = 1
	w := uint(7)
	for one<<w > 0 {
		w = 2*w + 1
	}
	return ^(one << w)
}

// MinOf returns the minimum value of the signed integer type T.
func MinOf[T ap.Integer]() T {
	return -MaxOf[T]() - 1
}

// Add returns a+b and false if the sum overflows T.
func Add[T ap.Integer](a, b T) (T, bool) {
	if b > 0 && a > MaxOf[T]()-b {
		return 0, false
	}
	if b < 0 && a < MinOf[T]()-b {
		return 0, false
	}
	return a + b, true
}

// Sub returns a-b and false if the difference overflows T.
func Sub[T ap.Integer](a, b T) (T, bool) {
	if b > 0 && a < MinOf[T]()+b {
		return 0, false
	}
	if b < 0 && a > MaxOf[T]()+b {
		return 0, false
	}
	return a - b, true
}

// AddWrap returns a+b wrapped into T modulo 2^w, where w is the width of T,
// as Go defines signed overflow, and the carry: +1 if the exact a+b is above
// MaxOf[T], −1 if it is below MinOf[T], and 0 if it fits, in which case sum
// is exact. So the exact a+b is sum + carry·2^w.
func AddWrap[T ap.Integer](a, b T) (sum T, carry int) {
	if b > 0 && a > MaxOf[T]()-b {
		carry = 1
	} else if b < 0 && a < MinOf[T]()-b {
		carry = -1
	}
	return a + b, carry
}

// magnitude returns |v| as a uint64, correct even for the minimum int64.
func magnitude(v int64) uint64 {
	if v < 0 {
		return -uint64(v)
	}
	return uint64(v)
}

// MulInt returns a*k and false if the product overflows T.
func MulInt[T ap.Integer](a T, k int) (T, bool) {
	if a == 0 || k == 0 {
		return 0, true
	}
	neg := (a < 0) != (k < 0)
	hi, lo := bits.Mul64(magnitude(int64(a)), magnitude(int64(k)))
	limit := uint64(MaxOf[T]())
	if neg {
		limit++
	}
	if hi != 0 || lo > limit {
		return 0, false
	}
	if neg {
		return T(-lo), true
	}
	return T(lo), true
}
