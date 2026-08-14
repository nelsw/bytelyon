package model

import (
	"time"

	"github.com/rs/zerolog"
)

type Bots []*Bot

type Bot struct {
	ID        int       `json:"id"`
	Type      BotType   `json:"type"`
	Query     string    `json:"query"`
	Headless  bool      `json:"headless"`
	LastRunAt time.Time `json:"last_run_at"`
	SitemapID int       `json:"sitemap_id,omitempty"`
	SearchID  int       `json:"serp_id,omitempty"`
}

func (b *Bot) MarshalZerologObject(evt *zerolog.Event) {
	evt.Int("#", b.ID).
		Str("q", b.Query).
		Any("t", b.Type)
}
