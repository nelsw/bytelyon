package util

import (
	"math/rand"
	"reflect"
)

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

func Eq[T comparable](a T, args ...T) bool {
	for _, b := range args {
		if a != b {
			return false
		}
	}
	return true
}
