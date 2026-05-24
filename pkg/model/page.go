package model

import (
	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/oklog/ulid/v2"
)

type Page struct {
	ID   ulid.ULID
	Data []byte
	URL  string
}

func NewPage(id ulid.ULID, data []byte, url string) *Page {
	return &Page{ID: id, Data: data, URL: url}
}

func (p *Page) Key() string {
	return util.Path("page", p.URL, p.ID)
}
