package news

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/page"
	"github.com/nelsw/bytelyon/pkg/s3"
	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/oklog/ulid/v2"
)

func key(userID ulid.ULID, topic string) string {
	return fmt.Sprintf("users/%s/news/%s.json", userID, topic)
}

func Delete(userID ulid.ULID, topic string) error {

	m, err := Find(userID, topic)
	if err != nil {
		return err
	}

	for url, h := range m {
		err = errors.Join(page.Delete(url, h.ID))
	}

	return errors.Join(s3.Delete(key(userID, topic), false))
}

func Find(userID ulid.ULID, topic string) (map[string]*Headline, error) {

	out, err := s3.Get(key(userID, topic), false)
	if err != nil {
		return nil, err
	}

	var m map[string]*Headline
	if err = json.Unmarshal(out, &m); err != nil {
		return nil, err
	}

	return m, nil
}

func FindOrNew(userID ulid.ULID, topic string) *model.SyncMap[string, *Headline] {
	m, err := Find(userID, topic)
	if err != nil {
		m = make(map[string]*Headline)
	}
	return model.NewSyncMap[string, *Headline](m)
}

func Save(userID ulid.ULID, topic string, entries map[string]*Headline) error {
	if m, err := Find(userID, topic); err == nil {
		for k, v := range m {
			entries[k] = v
		}
	}
	return s3.Put(key(userID, topic), util.JSON(entries), false)
}
