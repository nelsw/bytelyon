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
	if len(m) == 0 {
		return make([]K, 0)
	}
	return slices.Sorted(maps.Keys(m))
}

func (m Map[K, V]) Values() (v []V) {
	v = make([]V, 0, len(m))
	for _, k := range m.Keys() {
		v = append(v, m[k])
	}
	return
}

func (m Map[K, V]) Has(k K) (ok bool) {
	_, ok = m[k]
	return
}

func (m Map[K, V]) Get(k K) (v V, ok bool) {
	v, ok = m[k]
	return
}

func (m Map[K, V]) Put(k K, v V) V {
	m[k] = v
	return v
}

func (m Map[K, V]) Delete(k K) (existed bool) {
	if existed = m.Has(k); existed {
		delete(m, k)
	}
	return
}
