package store

import (
	"encoding/json"
	"fmt"
	"iter"
	"maps"
	"slices"
	"strings"
	"sync"

	"github.com/nelsw/bytelyon/pkg/s3"
)

type DB[K comparable, V any] struct {
	m         sync.Mutex
	table     map[K]V
	key       string
	committed bool
}

func New[K comparable, V any](args ...any) (*DB[K, V], error) {
	s := new(DB[K, V])
	s.committed = true
	s.table = make(map[K]V)

	var arr []string
	for _, a := range args {
		arr = append(arr, fmt.Sprint(a))
	}
	s.key = strings.Join(arr, "/")
	if !strings.HasSuffix(s.key, ".json") {
		s.key += ".json"
	}

	return s, s.init()
}

func (db *DB[K, V]) init() error {

	b, err := s3.GetPrivateObject(db.key)
	if err == nil {
		return json.Unmarshal(b, &db.table)
	}

	if strings.Contains(err.Error(), "StatusCode: 404") {
		db.table = make(map[K]V)
		return nil
	}

	return err
}

func (db *DB[K, V]) Close() error {
	return db.Commit()
}

func (db *DB[K, V]) Keys() []K {
	return slices.Collect(maps.Keys(db.table))
}

func (db *DB[K, V]) Values() []V {
	return slices.Collect(maps.Values(db.table))
}

func (db *DB[K, V]) All() iter.Seq2[K, V] {
	return maps.All(db.table)
}

func (db *DB[K, V]) Get(k K) V {
	return db.table[k]
}

func (db *DB[K, V]) Put(k K, v V) {
	db.table[k] = v
	db.committed = false
}

func (db *DB[K, V]) Drop(k K) {
	delete(db.table, k)
	db.committed = false
}

func (db *DB[K, V]) Commit() error {

	if db.committed {
		return nil
	}

	if b, err := json.Marshal(db.table); err != nil {
		return err
	} else if err = s3.PutPrivateObject(db.key, b); err != nil {
		return err
	}

	db.committed = true
	return nil
}
