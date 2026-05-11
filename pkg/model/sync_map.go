package model

import (
	"cmp"
	"sync"
)

// SyncMap is a thread-safe, orderable map.
type SyncMap[K cmp.Ordered, V any] struct {

	// Map is the underlying data structure that stores entries.
	Map[K, V]

	// mutex (mutual exclusion) prevents simultaneous data access.
	mutex sync.Mutex
}

func NewSyncMap[K cmp.Ordered, V any](m ...map[K]V) *SyncMap[K, V] {
	return &SyncMap[K, V]{
		Map: MakeMap[K, V](m...),
	}
}

func (m *SyncMap[K, V]) ToMap() Map[K, V] {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.Map
}

func (m *SyncMap[K, V]) Has(k K) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.Map.Has(k)
}

func (m *SyncMap[K, V]) Get(k K) (V, bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.Map.Get(k)
}

func (m *SyncMap[K, V]) Set(k K, v V) V {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.Map.Set(k, v)
}

func (m *SyncMap[K, V]) Delete(k K) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.Map.Delete(k)
}

func (m *SyncMap[K, V]) Keys() []K {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.Map.Keys()
}

func (m *SyncMap[K, V]) Values() []V {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.Map.Values()
}
