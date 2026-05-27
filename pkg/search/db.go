package search

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	"github.com/nelsw/bytelyon/pkg/page"
	"github.com/nelsw/bytelyon/pkg/s3"
	"github.com/nelsw/bytelyon/pkg/serp"
	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/oklog/ulid/v2"
)

func key(userID ulid.ULID, query string) string {
	return fmt.Sprintf("users/%s/search/%s.json", userID, query)
}

func Delete(userID ulid.ULID, query string) error {

	m, err := Find(userID, query)
	if err != nil {
		return err
	}

	for id, url := range m {
		err = errors.Join(page.Delete(url, id))
	}

	return errors.Join(s3.Delete(key(userID, query), false))
}

func Find(userID ulid.ULID, query string) (map[ulid.ULID]string, error) {

	out, err := s3.Get(key(userID, query), false)
	if err != nil {
		return nil, err
	}

	var m map[ulid.ULID]string
	if err = json.Unmarshal(out, &m); err != nil {
		return nil, err
	}

	return m, nil
}

func FindIDs(userID ulid.ULID, query string) ([]ulid.ULID, error) {
	m, err := Find(userID, query)
	if err != nil {
		return nil, err
	}
	var ids []ulid.ULID
	for id := range m {
		ids = append(ids, id)
	}
	slices.SortFunc(ids, func(a, b ulid.ULID) int {
		return b.Compare(a)
	})
	return ids, nil
}

func FindSerp(userID ulid.ULID, query string, id ulid.ULID) (*serp.Model, error) {
	m, err := Find(userID, query)
	if err != nil {
		return nil, err
	}

	url, ok := m[id]
	if !ok {
		return nil, nil
	}

	return page.FindObject[*serp.Model](url, id)
}

func Save(userID ulid.ULID, query string, m map[ulid.ULID]string) error {
	return s3.Put(key(userID, query), util.JSON(m), false)
}
