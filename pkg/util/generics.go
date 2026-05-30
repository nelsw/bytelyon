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

func Between[T int | float64](min, max T) T {
	return T(rand.Intn(int(max)-int(min)) + int(min))
}

// Or returns the first argument that is not zero, or the final argument if all are zero.
func Or[T any](ors ...T) (or T) {
	var v reflect.Value
	for _, or = range ors {
		if v = reflect.ValueOf(or); v.IsValid() && !v.IsZero() {
			return
		}
	}
	return
}
