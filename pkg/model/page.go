package model

import (
	"time"

	. "github.com/nelsw/bytelyon/pkg/util"
	"github.com/oklog/ulid/v2"
)

type Pages []*Page

type Page struct {
	ID        ulid.ULID `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	URL       string    `json:"url"`
	Domain    string    `json:"domain"`
	Path      string    `json:"path"`
	Title     string    `json:"title"`
	IMG       string    `json:"img"`
	HTML      string    `json:"html"`
	SERP      Serp      `json:"serp,omitempty"`
}

func NewPage(id ulid.ULID, url string) *Page {
	return &Page{
		ID:        id,
		CreatedAt: id.Timestamp().UTC(),
		URL:       url,
		Domain:    Domain(url),
		Path:      Path(url),
	}
}
