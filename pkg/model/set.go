package model

import (
	"cmp"
)

type Set[K cmp.Ordered] Map[K, bool]

func MakeSet[K cmp.Ordered](vals ...K) Set[K] {
	m := MakeMap[K, bool]()
	for _, v := range vals {
		m.Put(v, true)
	}
	return Set[K](m)
}

func (s Set[K]) Add(k K)    { Map[K, bool](s).Put(k, true) }
func (s Set[K]) Slice() []K { return Map[K, bool](s).Keys() }
