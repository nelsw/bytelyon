package snippet

import (
	"github.com/nelsw/bytelyon/pkg/id"
	"github.com/oklog/ulid/v2"
)

type Model struct {
	ID    ulid.ULID         `json:"id"`
	Meta  map[string]string `json:"meta"`
	Title string            `json:"title"`
	URL   string            `json:"url"`
}

func New(url, title string, meta map[string]string) *Model {
	return &Model{
		ID:    id.NewULID(),
		Meta:  meta,
		Title: title,
		URL:   url,
	}
}
