package contract

import "cmp"

type Entry[K cmp.Ordered] interface {
	Key() K
	Value() []byte
}
