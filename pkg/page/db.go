package page

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/nelsw/bytelyon/pkg/s3"
	"github.com/nelsw/bytelyon/pkg/urls"
	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/oklog/ulid/v2"
)

func key(url string, id ulid.ULID, name string) string {
	return util.Path("pages", urls.PR(url), id, name)
}

func save(url string, id ulid.ULID, data []byte, name string) error {
	return s3.Put(key(url, id, name), data, strings.HasSuffix(name, ".png"))
}

func Delete(url string, id ulid.ULID) error {
	var err error
	err = errors.Join(s3.Delete(key(url, id, "object.json"), false))
	err = errors.Join(s3.Delete(key(url, id, "screenshot.png"), true))
	err = errors.Join(s3.Delete(key(url, id, "content.html"), false))
	return err
}

func FindObjects[T any](url string) (out []T, err error) {

	prefix := util.Path("pages", urls.PR(url))

	var keys []string
	if keys, err = s3.ListDirectories(prefix); err != nil {
		return nil, err
	}

	var id ulid.ULID
	var t T
	for _, k := range keys {
		s := strings.TrimPrefix(k, prefix+"/")
		s = strings.TrimSuffix(s, "/object.json")
		if id, _ = ulid.Parse(s); id.IsZero() {
			continue
		}
		if t, err = FindObject[T](url, id); err == nil {
			out = append(out, t)
		}
	}
	return
}

func FindObject[T any](url string, id ulid.ULID) (t T, err error) {
	var out []byte
	if out, err = s3.Get(key(url, id, "object.json"), false); err == nil {
		err = json.Unmarshal(out, &t)
	}
	return
}

func SaveObject(url string, id ulid.ULID, a any) error {
	return save(url, id, util.JSON(a), "object.json")
}

func SaveScreenshot(url string, id ulid.ULID, b []byte) error {
	return save(url, id, b, "screenshot.png")
}

func SaveContent(url string, id ulid.ULID, s string) error {
	return save(url, id, []byte(s), "content.html")
}
