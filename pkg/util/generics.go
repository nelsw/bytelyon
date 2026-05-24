package util

import (
	"math/rand"
	"reflect"

	"github.com/rs/zerolog/log"
)

// Ptr returns a pointer reference of the given value.
func Ptr[T any](t T) *T { return &t }

// Safe returns only the first value of an errorable ƒ.
func Safe[T any](t T, err error) T {
	if err != nil {
		log.Warn().Msgf("Suppressed: %v", err)
	}
	return t
}

func PtrOrNil[T any](a T) *T {
	if v := reflect.ValueOf(a); v.IsZero() || v.IsNil() {
		return nil
	}
	return &a
}

func Between[T int | float64](min, max T) T {
	return T(rand.Intn(int(max)-int(min)) + int(min))
}

// Or returns the first argument that is not zero, or the final argument if all are zero.
func Or[T any](ors ...T) T {
	return OrFunc(func(or T) bool { return true }, ors...)
}

// OrFunc returns the first argument that is not zero and returns true from the given function;
// else the final argument is returned.
func OrFunc[T any](f func(or T) bool, ors ...T) T {
	var or T
	var v reflect.Value
	for _, or = range ors {
		if v = reflect.ValueOf(or); v.IsValid() && !v.IsZero() && f(or) {
			return or
		}
	}
	return or
}
