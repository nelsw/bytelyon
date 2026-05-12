package store

import (
	"cmp"
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/s3"
	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/rs/zerolog/log"
)

type DB[K cmp.Ordered, V any] struct {
	table     *model.SyncMap[K, V]
	key       string
	committed bool
}

func New[K cmp.Ordered, V any](args ...any) (*DB[K, V], error) {

	s := new(DB[K, V])
	s.committed = true

	var arr []string
	for _, a := range args {
		arr = append(arr, fmt.Sprint(a))
	}
	s.key = strings.Join(arr, "/")

	// if we're missing a file extension, assume json
	if path.Ext(s.key) == "" {
		s.key += ".json"
	}

	l := log.With().Str("store", s.key).Logger()

	l.Trace().Send()

	b, err := s3.GetPrivateObject(s.key)
	if err == nil {
		var m map[K]V
		_ = json.Unmarshal(b, &m)
		s.table = model.NewSyncMap[K, V](m)
	} else if strings.Contains(err.Error(), "StatusCode: 404") {
		s.table = model.NewSyncMap[K, V]()
	}

	if err != nil {
		l.Err(err).Send()
		return nil, err
	}

	l.Debug().Send()

	return s, err
}

func (db *DB[K, V]) String() string {
	return string(util.JSON(map[string]any{
		"key":   db.key,
		"table": db.table,
	}))
}

func (db *DB[K, V]) init() error {

	b, err := s3.GetPrivateObject(db.key)
	if err == nil {
		return json.Unmarshal(b, &db.table)
	}

	if strings.Contains(err.Error(), "StatusCode: 404") {
		db.table = model.NewSyncMap[K, V]()
		return nil
	}

	return err
}

func (db *DB[K, V]) Close() error {
	return db.Commit()
}

func (db *DB[K, V]) Keys() []K {
	return db.table.Keys()
}

func (db *DB[K, V]) Values() []V {
	return db.table.Values()
}

func (db *DB[K, V]) Get(k K) (V, bool) {
	return db.table.Get(k)
}

func (db *DB[K, V]) Put(k K, v V) {
	db.table.Set(k, v)
	db.committed = false
}

func (db *DB[K, V]) Set(k K, v V) V {
	db.table.Set(k, v)
	db.committed = false
	return v
}

func (db *DB[K, V]) Drop(k K) {
	db.table.Delete(k)
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
