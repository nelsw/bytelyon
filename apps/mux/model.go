package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type BotType string

func (t *BotType) UnmarshalJSON(payload []byte) error {
	if text := string(payload); text == `"news"` || text == `"search"` || text == `"sitemap"` {
		*t = BotType(strings.ReplaceAll(text, `"`, ""))
		return nil
	}
	return fmt.Errorf("unknown bot type: %s", payload)
}

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

func (b *Bot) Run() {
	log.Info().EmbedObject(b).Msg("POST")
	Post[any](os.Getenv("BRO_URL")+"/bots", b)
}

func Bots() (bots []*Bot) {
	bots = Get[[]*Bot](os.Getenv("WEB_URL") + "/api/bots")
	log.Info().Int("size", len(bots)).Msg("GET")
	return
}
