package model

import (
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

type BotType string

const (
	NewsBot    BotType = "news"
	SearchBot  BotType = "search"
	SitemapBot BotType = "sitemap"
)

func (t *BotType) String() string {
	return string(*t)
}

func (t *BotType) UnmarshalJSON(payload []byte) error {
	if text := string(payload); text == `"news"` || text == `"search"` || text == `"sitemap"` {
		*t = BotType(strings.ReplaceAll(text, `"`, ""))
		return nil
	}
	return fmt.Errorf("unknown bot type: %s", payload)
}

type Bots []*Bot

type Bot struct {
	ID        int       `json:"id"`
	Type      BotType   `json:"type"`
	Query     string    `json:"query"`
	Headless  bool      `json:"headless"`
	Since     time.Time `json:"last_run_at"`
	SitemapID int       `json:"sitemap_id,omitempty"`
	SerpID    int       `json:"serp_id,omitempty"`
}

func (b *Bot) MarshalZerologObject(evt *zerolog.Event) {
	evt.Int("#", b.ID).
		Str("q", b.Query).
		Any("t", b.Type)
}

type Meta map[string][]string

type Pages []Page
type Page struct {
	Domain string `json:"domain"`
	Key    string `json:"screenshot_key"`
	Meta   Meta   `json:"meta,omitempty"`
	Title  string `json:"title"`
	URL    string `json:"url"`
	Kind   string `json:"kind"`
	Index  int    `json:"index"`
	Img    []byte `json:"screenshot"`
}

// todo - custom json unmarshal for serp img keyu and page img keys
type Serp struct {
	Query             string   `json:"query"`
	Img               []byte   `json:"screenshot"`
	Src               []byte   `json:"content"`
	Data              any      `json:"data"`
	SponsoredProducts Pages    `json:"sponsored_products"`
	SponsoredResults  Pages    `json:"sponsored_results"`
	OrganicResults    Pages    `json:"organic_results"`
	OrganicProducts   Pages    `json:"organic_products"`
	SimilarQueries    []string `json:"similar_queries"`
}

func (s Serp) Pages() (pages Pages) {
	pages = append(pages, s.SponsoredProducts...)
	pages = append(pages, s.SponsoredResults...)
	pages = append(pages, s.OrganicResults...)
	pages = append(pages, s.OrganicProducts...)
	return
}

// todo - custom json unmarshal for keys
type Sitemap struct {
	Domain string   `json:"domain"`
	URLs   []string `json:"urls"`
	Pages  Pages    `json:"pages,omitempty"`
}
