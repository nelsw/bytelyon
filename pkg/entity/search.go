package entity

import (
	"encoding/json"

	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/oklog/ulid/v2"
)

type Search struct {
	UTC  uint64
	Data []any
	*model.Bot
	model.Map[string, ulid.ULID]
}

func NewSearch(bot *model.Bot) *Search {
	return &Search{
		Bot: bot,
		UTC: ulid.Now(),
		Map: model.MakeMap[string, ulid.ULID](),
	}
}

func (s *Search) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"bot":  s.Bot,
		"data": s.Data,
	})
}

func (s *Search) Add(p *Page) {
	s.Map.Set(p.URL, p.ID)
	s.Data = append(s.Data, p)
}
