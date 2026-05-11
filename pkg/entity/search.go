package entity

import (
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/oklog/ulid/v2"
)

type Search struct {
	ID    ulid.ULID `json:"id"`
	Query string    `json:"query"`
	URLs  []string  `json:"urls"`

	Pages *model.SyncMap[string, *Page] `json:"-"`
}

func NewSearch(query string) *Search {
	return &Search{
		ID:    model.NewULID(),
		Query: query,
		Pages: model.NewSyncMap[string, *Page](),
	}
}

func (s *Search) AddPage(p *Page) {
	s.URLs = append(s.URLs, p.URL)
	s.Pages.Set(p.URL, p)
}
