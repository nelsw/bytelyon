package entity

import (
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/oklog/ulid/v2"
)

type News struct {
	ID    ulid.ULID `json:"id"`
	Topic string    `json:"topic"`
	URLs  []string  `json:"urls"`

	Pages *model.SyncMap[string, *Page] `json:"-"`
}

func NewNews(topic string) *News {
	return &News{
		ID:    model.NewULID(),
		Topic: topic,
		Pages: model.NewSyncMap[string, *Page](),
	}
}

func (n *News) AddPage(p *Page) {
	n.URLs = append(n.URLs, p.URL)
	n.Pages.Set(p.URL, p)
}
