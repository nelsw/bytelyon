package util

import (
	"math/rand"
	"strings"
)

func Ptr[T any](a T) *T { return &a }

func Between[T int | float64](min, max T) T {
	return T(rand.Intn(int(max)-int(min)) + int(min))
}

func Domain(s string) string {
	s = strings.TrimPrefix(s, "https://")
	s = strings.TrimPrefix(s, "http://")
	s = strings.TrimPrefix(s, "www.")
	s = strings.Split(s, "/")[0]
	for strings.Count(s, ".") > 1 {
		s = strings.Split(s, ".")[1]
	}
	return s
}

func Capitalize(s string) string {
	return strings.ToUpper(s[0:1]) + s[1:]
}
