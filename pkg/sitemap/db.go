package sitemap

import (
	"fmt"
	"maps"
	"slices"

	"github.com/nelsw/bytelyon/pkg/s3"
	"github.com/nelsw/bytelyon/pkg/util/json"
	"github.com/oklog/ulid/v2"
)

func key(userID ulid.ULID, domain string) string {
	return fmt.Sprintf("users/%s/sitemap/%s/result.json", userID, domain)
}

func Delete(userID ulid.ULID, domain string) error { return s3.Delete(key(userID, domain), false) }

func Find(userID ulid.ULID, domain string) (arr []string) {
	if out, err := s3.Get(key(userID, domain), false); err == nil {
		arr = json.To[[]string](out)
	}
	return
}

func Save(userID ulid.ULID, domain string, m map[string]bool) error {
	for _, url := range Find(userID, domain) {
		m[url] = true
	}
	arr := slices.Sorted(maps.Keys(m))
	return s3.Put(key(userID, domain), json.Of(arr), false)
}
