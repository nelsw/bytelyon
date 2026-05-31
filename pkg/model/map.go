package model

import (
	"cmp"
	"maps"
	"sync"
)

// SyncMap is a thread-safe, orderable map.
type SyncMap[K cmp.Ordered, V any] struct {
	m map[K]V
	sync.RWMutex
}

func NewSyncMap[K cmp.Ordered, V any](m ...map[K]V) *SyncMap[K, V] {
	var sm = new(SyncMap[K, V])
	if len(m) == 0 {
		sm.m = make(map[K]V)
	} else {
		sm.m = m[0]
	}
	return sm
}

func (m *SyncMap[K, V]) Clone() map[K]V {
	m.Lock()
	defer m.Unlock()
	return maps.Clone(m.m)
}

func (m *SyncMap[K, V]) Drop(k K) {
	m.Lock()
	defer m.Unlock()
	delete(m.m, k)
}

func (m *SyncMap[K, V]) Has(k K) (ok bool) {
	m.Lock()
	defer m.Unlock()
	_, ok = m.m[k]
	return
}

func (m *SyncMap[K, V]) Put(k K, v V) {
	m.Lock()
	defer m.Unlock()
	m.m[k] = v
}
