package model

import (
	"encoding/json"

	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/oklog/ulid/v2"
)

type Page struct {
	ID    ulid.ULID `json:"id"`
	URL   string    `json:"url"`
	Title string    `json:"title"`
	Meta  *Meta     `json:"meta,omitempty"`
	SERP  *Serp     `json:"serp,omitempty"`
}

func NewPage(id ulid.ULID, url, title string) *Page {
	return &Page{
		ID:    id,
		URL:   url,
		Title: title,
	}
}

func (p *Page) SetSerp(content string) {
	p.SERP = util.Ptr(MakeSerp(p.URL, content))
}

func (p *Page) SetMeta(content string) {
	p.Meta = util.Ptr(ParseMeta(content))
}

func (p *Page) String() string {
	b, _ := json.MarshalIndent(p, "", "\t")
	return string(b)
}
