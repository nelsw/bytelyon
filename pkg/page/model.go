package page

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/s3"
	"github.com/nelsw/bytelyon/pkg/urls"
	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/oklog/ulid/v2"
)

type Model struct {
	id   ulid.ULID
	data []byte
	url  string
}

func New(url string, data any, id ...ulid.ULID) *Model {

	var m = new(Model)
	m.url = url

	if len(id) == 0 {
		m.id = model.NewULID()
	} else {
		m.id = id[0]
	}

	switch v := data.(type) {
	case string:
		m.data = []byte(v)
	case []byte:
		m.data = v
	default:
		m.data = util.JSON(data)
	}

	return m
}

func Create(url string, id ulid.ULID, data ...any) (err error) {
	for _, d := range data {
		err = errors.Join(err, New(url, d, id).save())
	}
	return
}

func Delete(u string, id ulid.ULID) error {
	prefix := util.Path("page", urls.PR(u), id)
	keys, err := s3.ListDirectories(prefix)
	if err != nil {
		return err
	}
	for _, k := range keys {
		s3.Delete(k, true)
	}
	return nil
}

func Find(u string, i ulid.ULID) (*Model, error) {

	key := util.Path("page", urls.PR(u), i, "object.json")

	out, err := s3.Get(key, false)
	if err != nil {
		return nil, err
	}

	m := new(Model)
	if err = json.Unmarshal(out, m); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *Model) save() error {

	var base string
	var isPublic bool
	if t := http.DetectContentType(m.data); strings.HasPrefix(t, "text/html") {
		base = "content.html"
	} else if strings.HasPrefix(t, "image/") {
		base = "screenshot.png"
		isPublic = true
	} else if json.Valid(m.data) {
		base = "object.json"
	} else {
		base = "unknown.txt"
	}

	key := util.Path("pages", urls.PR(m.url), m.id, base)

	return s3.Put(key, m.data, isPublic)
}
