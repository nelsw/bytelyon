package model

import (
	"encoding/json"
	"fmt"

	"github.com/nelsw/bytelyon/apps/mux/internal/doc"
	"github.com/nelsw/bytelyon/apps/mux/internal/util/urls"
	"github.com/nelsw/bytelyon/apps/mux/internal/util/uuid"
	"github.com/rs/zerolog/log"
)

type Page struct {
	Domain        string              `json:"domain"`
	Meta          map[string][]string `json:"meta"`
	ScreenshotKey string              `json:"screenshot_key"`
	Title         string              `json:"title"`
	URL           string              `json:"url"`
	Index         int                 `json:"index"`
	Kind          string              `json:"kind"`
	screenshot    []byte
	route         string
}

func (p *Page) Route() string {
	return p.route
}

func (p *Page) Screenshot() (string, []byte, bool) {
	return p.URL, p.screenshot, true
}

func (p *Page) UnmarshalJSON(data []byte) (err error) {
	defer func() {
		if err != nil {
			log.Err(err).Bytes("data", data).Msg("Page:UnmarshalJSON")
		}
	}()

	var v struct {
		BotID      int     `json:"bot_id"`
		BotType    BotType `json:"bot_type"`
		Content    []byte  `json:"content"`
		Index      int     `json:"index"`
		Kind       string  `json:"kind"`
		Screenshot []byte  `json:"screenshot"`
		SearchID   int     `json:"search_id"`
		SitemapID  int     `json:"sitemap_id"`
		Title      string  `json:"title"`
		URL        string  `json:"url"`
	}

	if err = json.Unmarshal(data, &v); err != nil {
		return
	}

	page := Page{
		Domain:        urls.Domain(v.URL),
		Meta:          doc.New(v.Content, true).Meta(),
		ScreenshotKey: fmt.Sprintf("%s/%d/%s/screenshot.png", v.BotType, v.BotID, uuid.FromURL(v.URL)),
		Title:         v.Title,
		URL:           v.URL,
		screenshot:    v.Screenshot,
	}

	if v.BotType == SitemapBot {
		page.route = fmt.Sprintf("/sitemaps/%d/page", v.SitemapID)
	} else {
		page.route = fmt.Sprintf("/searches/%d/page", v.SearchID)
		page.Index = v.Index
		page.Kind = v.Kind
	}

	*p = page

	return
}
