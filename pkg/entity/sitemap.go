package entity

import (
	"encoding/json"

	"github.com/nelsw/bytelyon/pkg/dto"
	"github.com/nelsw/bytelyon/pkg/model"
)

type Sitemap struct {
	*model.Bot
	*model.SyncMap[string, *Page]
	root *dto.Node
}

func NewSitemap(bot *model.Bot) *Sitemap {
	return &Sitemap{
		Bot:     bot,
		SyncMap: model.NewSyncMap[string, *Page](),
		root:    dto.NewNode("https://" + bot.Target),
	}
}

func (s *Sitemap) Merge(x *Sitemap) {
	for _, p := range x.Values() {
		s.Add(p)
	}
}

func (s *Sitemap) Add(p *Page) {
	s.Set(p.URL, p)
	s.root.Add(p.URL)
}

func (s *Sitemap) MarshalJSON() ([]byte, error) {
	for _, url := range s.Keys() {
		s.root.Add(url)
	}
	return json.Marshal(map[string]any{
		"bot":   s.Bot,
		"nodes": s.root.Children.Values(),
	})
}
