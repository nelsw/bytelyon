package model

import (
	"cmp"
)

type Set[K cmp.Ordered] struct {
	m Map[K, bool]
}

func NewSet[K cmp.Ordered](vals ...K) *Set[K] {
	var s = new(Set[K])
	s.m = MakeMap[K, bool]()
	for _, v := range vals {
		s.m.Set(v, true)
	}
	return s
}

func (s *Set[K]) Add(k K)    { s.m.Set(k, true) }
func (s *Set[K]) Slice() []K { return s.m.Keys() }
