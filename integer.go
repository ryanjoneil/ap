package ap

// Integer is any native signed int type.
type Integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}
