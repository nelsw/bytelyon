package model

import (
	"encoding/json"
	"slices"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type Set struct {
	d Data[bool]
	m sync.Mutex
}

func (s *Set) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Slice())
}

func (s *Set) ToAttributeValue() *types.AttributeValueMemberSS {
	return &types.AttributeValueMemberSS{Value: s.Slice()}
}

func (s *Set) Drop(key string, lock ...bool) {
	if len(lock) > 0 && lock[0] {
		s.m.Lock()
		defer s.m.Unlock()
	}
	s.d.Delete(key)
}

func (s *Set) Add(a any, lock ...bool) {
	if a == nil {
		return
	}
	if len(lock) > 0 && lock[0] {
		s.m.Lock()
		defer s.m.Unlock()
	}
	switch a.(type) {
	case *Set:
		for _, v := range a.(*Set).Slice() {
			s.d.Set(v, true)
		}
	case []any:
		for _, v := range a.([]any) {
			s.d.Set(v.(string), true)
		}
	case []string:
		for _, v := range a.([]string) {
			s.d.Set(v, true)
		}
	case *types.AttributeValueMemberSS:
		for _, v := range a.(*types.AttributeValueMemberSS).Value {
			s.d.Set(v, true)
		}
	default:
		s.d.Set(a.(string), true)
	}
}

func (s *Set) Has(key string, lock ...bool) bool {
	if len(lock) > 0 && lock[0] {
		s.m.Lock()
		defer s.m.Unlock()
	}
	return s.d.Has(key)
}

func (s *Set) Len(lock ...bool) int {
	if len(lock) > 0 && lock[0] {
		s.m.Lock()
		defer s.m.Unlock()
	}
	return s.d.Len()
}

func (s *Set) Slice(lock ...bool) []string {
	if len(lock) > 0 && lock[0] {
		s.m.Lock()
		defer s.m.Unlock()
	}
	v := s.d.Keys()
	slices.Sort(v)
	return v
}

func ParseSet(a any) *Set {
	s := NewSet()
	s.Add(a)
	return s
}

func NewSet() *Set {
	return &Set{d: MakeData[bool]()}
}
