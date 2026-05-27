package news

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	"github.com/nelsw/bytelyon/pkg/page"
	"github.com/nelsw/bytelyon/pkg/s3"
	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/oklog/ulid/v2"
)

func key(userID ulid.ULID, topic string) string {
	return fmt.Sprintf("users/%s/news/%s.json", userID, topic)
}

func Delete(userID ulid.ULID, topic string) (err error) {
	for _, h := range Find(userID, topic) {
		err = errors.Join(err, page.Delete(h.URL, h.ID))
	}
	return errors.Join(err, s3.Delete(key(userID, topic), false))
}

func Find(userID ulid.ULID, topic string) (arr []*Model) {
	if out, err := s3.Get(key(userID, topic), false); err == nil {
		err = json.Unmarshal(out, &arr)
	}
	return
}

func Save(userID ulid.ULID, topic string, arr []*Model) error {
	slices.SortFunc(arr, func(a, b *Model) int { return b.ID.Compare(a.ID) })
	return s3.Put(key(userID, topic), util.JSON(arr), false)
}
