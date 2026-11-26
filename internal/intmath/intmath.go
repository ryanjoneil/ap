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

// FitsScaled reports whether k*|c| <= MaxOf[T]. It returns false if k <= 0 or
// c == MinOf[T], since |MinOf[T]| is not representable.
func FitsScaled[T ap.Integer](k int, c T) bool {
	if k <= 0 || c == MinOf[T]() {
		return false
	}
	return magnitude(int64(c)) <= uint64(MaxOf[T]())/uint64(k)
}
