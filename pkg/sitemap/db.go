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
	return fmt.Sprintf("users/%s/sitemap/%s.json", userID, domain)
}

func Delete(userID ulid.ULID, domain string) error {
	return s3.Delete(key(userID, domain), false)
}

func Find(userID ulid.ULID, domain string) (arr []string) {
	if out, err := s3.Get(key(userID, domain), false); err == nil {
		err = json.Unmarshal(out, &arr)
	}
	return
}

func Save(userID ulid.ULID, domain string, urls *model.SyncMap[string, bool]) error {
	set := model.NewSet[string](Find(userID, domain)...)
	for k, v := range urls.Map {
		if v {
			set.Add(k)
		}
	}
	return s3.Put(key(userID, domain), util.JSON(set.Slice()), false)
}
