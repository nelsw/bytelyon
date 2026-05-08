package store

import (
	"encoding/json"
	"iter"
	"maps"
	"slices"
	"strings"

	"github.com/nelsw/bytelyon/pkg/s3"
)

type table[T any] map[string]T

type DB[T any] struct {
	table table[T]
	key   string
}

func New[T any](key string) (*DB[T], error) {
	s := new(DB[T])
	s.table = make(map[string]T)
	s.key = key
	return s, s.init()
}

func (db *DB[T]) init() error {

	b, err := s3.GetPrivateObject(db.key)
	if err == nil {
		return json.Unmarshal(b, &db.table)
	}

	if strings.Contains(err.Error(), "StatusCode: 404") {
		db.table = make(map[string]T)
		return nil
	}

	return err
}

func (db *DB[T]) Close() error {
	return db.Commit()
}

func (db *DB[T]) Keys() []string {
	return slices.Collect(maps.Keys(db.table))
}

func (db *DB[T]) All() iter.Seq2[string, T] {
	return maps.All(db.table)
}

func (db *DB[T]) Get(k string) T {
	return db.table[k]
}

func (db *DB[T]) Put(k string, v T) {
	db.table[k] = v
}

func (db *DB[T]) Commit() error {
	b, err := json.Marshal(db.table)
	if err != nil {
		return err
	}
	return s3.PutPrivateObject(db.key, b)
}
