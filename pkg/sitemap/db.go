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

func Find(userID ulid.ULID, domain string) ([]string, error) {

	out, err := s3.Get(key(userID, domain), false)
	if err != nil {
		return nil, err
	}

	var arr []string
	if err = json.Unmarshal(out, &arr); err != nil {
		return nil, err
	}

	return arr, nil
}

func Save(userID ulid.ULID, domain string, urls *model.SyncMap[string, bool]) error {

	set := model.NewSet[string]()
	for k, v := range urls.Map {
		if v {
			set.Add(k)
		}
	}

	if arr, err := Find(userID, domain); err == nil {
		for _, url := range arr {
			set.Add(url)
		}
	}
	return s3.Put(key(userID, domain), util.JSON(set.Slice()), false)
}
