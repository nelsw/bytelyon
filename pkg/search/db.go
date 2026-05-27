package search

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/nelsw/bytelyon/pkg/page"
	"github.com/nelsw/bytelyon/pkg/s3"
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

	for id, urls := range m {
		for _, url := range urls {
			err = errors.Join(page.Delete(url, id))
		}
	}

	return errors.Join(s3.Delete(key(userID, query), false))
}

func Find(userID ulid.ULID, query string) (map[ulid.ULID][]string, error) {

	out, err := s3.Get(key(userID, query), false)
	if err != nil {
		return nil, err
	}

	var m map[ulid.ULID][]string
	if err = json.Unmarshal(out, &m); err != nil {
		return nil, err
	}

	return m, nil
}

func Save(userID ulid.ULID, query string, m map[ulid.ULID][]string) error {
	if f, err := Find(userID, query); err == nil {
		for k, v := range f {
			m[k] = v
		}
	}
	return s3.Put(key(userID, query), util.JSON(m), false)
}
