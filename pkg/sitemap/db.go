package sitemap

import (
	"encoding/json"
	"fmt"

	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/s3"
	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/oklog/ulid/v2"
)

func key(userID ulid.ULID, domain string) string {
	return fmt.Sprintf("users/%s/sitemap/%s/result.json", userID, domain)
}

func Delete(userID ulid.ULID, domain string) error { return s3.Delete(key(userID, domain), false) }

func Find(userID ulid.ULID, domain string) (arr []string) {
	if out, err := s3.Get(key(userID, domain), false); err == nil {
		err = json.Unmarshal(out, &arr)
	}
	return
}

func Save(userID ulid.ULID, domain string, urls *model.SyncMap[string, bool]) error {
	for _, url := range Find(userID, domain) {
		urls.Set(url, true)
	}
	return s3.Put(key(userID, domain), util.JSON(urls.Keys()), false)
}
