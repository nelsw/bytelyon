package model

import (
	"cmp"
	"slices"
)

type SyncSet[K cmp.Ordered] struct {
	*SyncMap[K, bool]
}

func NewSyncSet[K cmp.Ordered]() *SyncSet[K] {
	return &SyncSet[K]{
		SyncMap: NewSyncMap[K, bool](),
	}
}

func (s *SyncSet[K]) Slice(truthy ...bool) []K {

	t := len(truthy) > 0

	var arr []K
	for k, v := range s.Map {
		if t && !v {
			continue
		}
		arr = append(arr, k)
	}
	slices.Sort(arr)
	return arr
}
