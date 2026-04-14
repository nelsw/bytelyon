package model

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type Data[T any] map[string]T

func (d Data[T]) String() string {
	b, _ := json.MarshalIndent(d, "", "\t")
	return string(b)
}
func (d Data[T]) Set(k string, value T) { d[k] = value }
func (d Data[T]) Get(k string) T        { return d[k] }
func (d Data[T]) Has(k string) bool     { _, ok := d[k]; return ok }
func (d Data[T]) Delete(k string)       { delete(d, k) }
func (d Data[T]) Len() int              { return len(d) }
func (d Data[T]) Keys() []string        { return slices.Collect(maps.Keys(d)) }
func (d Data[T]) ToAttributeValue() *types.AttributeValueMemberM {
	m := make(map[string]types.AttributeValue)
	for _, k := range d.Keys() {
		m[k] = &types.AttributeValueMemberS{Value: fmt.Sprint(d.Get(k))}
	}
	return &types.AttributeValueMemberM{Value: m}
}

func ParseData[T any](a any) Data[T] {

	if a == nil {
		return nil
	}

	d := make(map[string]T)

	switch a.(type) {

	case *types.AttributeValueMemberM:
		var s any
		for k, v := range a.(*types.AttributeValueMemberM).Value {
			s = v.(*types.AttributeValueMemberS).Value
			if t, ok := s.(T); ok {
				d[k] = t
			}
		}
	}

	return d
}

func MakeData[T any]() Data[T] { return make(map[string]T) }
