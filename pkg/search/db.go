package search

import (
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/nelsw/bytelyon/pkg/s3"
	"github.com/nelsw/bytelyon/pkg/serp"
	"github.com/nelsw/bytelyon/pkg/util/json"
	"github.com/oklog/ulid/v2"
)

func key(userID ulid.ULID, query string) string {
	return fmt.Sprintf("users/%s/search/%s/result.json", userID, query)
}

func Delete(userID ulid.ULID, query string) error {

	arr, err := Find(userID, query)
	if err == nil {
		for _, id := range arr {
			err = errors.Join(err, serp.Delete(query, id))
		}
	}

	return errors.Join(err, s3.Delete(key(userID, query), false))
}

func Find(userID ulid.ULID, query string) (arr []ulid.ULID, err error) {
	var out []byte
	if out, err = s3.Get(key(userID, query), false); err != nil {
		return
	}
	arr = json.To[[]ulid.ULID](out)
	slices.SortFunc(arr, func(a, b ulid.ULID) int { return b.Compare(a) })
	return
}

func FindSerp(userID ulid.ULID, query string, id ulid.ULID) (*serp.Model, error) {
	arr, err := Find(userID, query)
	if err != nil {
		return nil, err
	}
	fmt.Println(arr)
	for _, a := range arr {
		if a.Compare(id) == 0 {
			return serp.Find(query, a)
		}
	}

	return nil, nil
}

func Save(userID ulid.ULID, query string, arr []ulid.ULID) error {
	return s3.Put(key(userID, query), json.Of(arr), false)
}

func Update(userID ulid.ULID, query string, id ulid.ULID) error {

	arr, _ := Find(userID, query)
	if len(arr) == 0 {
		return Save(userID, query, []ulid.ULID{id})
	}

	m := map[ulid.ULID]bool{id: true}
	for _, a := range arr {
		m[a] = true
	}
	arr = slices.Collect(maps.Keys(m))
	slices.SortFunc(arr, func(a, b ulid.ULID) int { return b.Compare(a) })
	return Save(userID, query, arr)
}
