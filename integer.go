package ap

import "unsafe"

// Integer is any native signed int type.
type Integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

// MaxOf returns the maximum value of any signed integer type.
func MaxOf[T Integer]() T {
	return ^MinOf[T]()
}

// MinOf returns the minimum value of any signed integer type.
func MinOf[T Integer]() T {
	// See: https://github.com/golang/go/issues/50019#issuecomment-1327464505
	var zero T
	minusone := ^zero
	bits := 8 * unsafe.Sizeof(zero)
	return minusone << (bits - 1)
}
