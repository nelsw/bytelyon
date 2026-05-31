package news

import (
	"errors"
	"fmt"

	"github.com/nelsw/bytelyon/pkg/page"
	"github.com/nelsw/bytelyon/pkg/s3"
	"github.com/nelsw/bytelyon/pkg/util/json"
	"github.com/oklog/ulid/v2"
)

func key(userID ulid.ULID, topic string) string {
	return fmt.Sprintf("users/%s/news/%s/result.json", userID, topic)
}

func Delete(userID ulid.ULID, topic string) (err error) {
	for _, h := range Find(userID, topic) {
		err = errors.Join(err, page.Delete(h.URL, h.ID))
	}
	return errors.Join(err, s3.Delete(key(userID, topic), false))
}

func Find(userID ulid.ULID, topic string) (arr []Model) {
	if out, err := s3.Get(key(userID, topic), false); err == nil {
		arr = json.To[[]Model](out)
	}
	return
}

func Save(userID ulid.ULID, topic string, arr []Model) error {
	return s3.Put(key(userID, topic), json.Of(arr), false)
}
