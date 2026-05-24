package snippet

import (
	"github.com/nelsw/bytelyon/pkg/document"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/urls"
	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/oklog/ulid/v2"
)

type Model struct {
	ID         ulid.ULID  `json:"id"`
	URL        string     `json:"url"`
	Title      string     `json:"title"`
	Meta       model.Meta `json:"meta"`
	Links      []string   `json:"links"`
	content    string
	screenshot []byte
}

func New(id ulid.ULID, url, content string, screenshot []byte) *Model {
	doc := document.New(content)
	return &Model{
		ID:         id,
		URL:        urls.PR(url),
		Title:      util.Or(doc.Title, doc.Meta.Title()),
		Meta:       doc.Meta,
		Links:      doc.Links,
		content:    content,
		screenshot: screenshot,
	}
}
