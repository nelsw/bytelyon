package model

import (
	"cmp"
	"maps"
	"slices"
)

type Map[K cmp.Ordered, V any] map[K]V

func MakeMap[K cmp.Ordered, V any](m ...map[K]V) Map[K, V] {
	if len(m) > 0 {
		return m[0]
	}
	return make(Map[K, V])
}

func (m Map[K, V]) Keys() []K {
	return slices.Sorted(maps.Keys(m))
}

func (m Map[K, V]) Values() (v []V) {
	for _, val := range m.Keys() {
		v = append(v, m[val])
	}
	return
}

func (m Map[K, V]) Has(k K) bool {
	_, ok := m[k]
	return ok
}

func (m Map[K, V]) Get(k K) (v V, ok bool) {
	v, ok = m[k]
	return
}

func (m Map[K, V]) Set(k K, v V) V {
	m[k] = v
	return v
}

func (m Map[K, V]) Delete(k K) bool {
	existed := m.Has(k)
	delete(m, k)
	return existed
}

func (m Map[K, V]) Len() int {
	return len(m)
}
