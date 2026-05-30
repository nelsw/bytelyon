package ptr

import "reflect"

var (
	True        = Of(true)
	False       = Of(false)
	ZeroFloat64 = Of(0.0)
)

// Of returns a pointer to the given value.
func Of[T any](t T) *T {
	return &t
}

// OrNil returns a pointer to the given value if it is not nil, valid, and not zero; else nil;
func OrNil[T any](t T) *T {
	if v := reflect.ValueOf(t); !v.IsNil() && v.IsValid() && !v.IsZero() {
		return Of(t)
	}
	return nil
}
