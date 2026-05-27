package search

import (
	"github.com/nelsw/bytelyon/internal/pw"
	"github.com/nelsw/bytelyon/pkg/page"
	"github.com/nelsw/bytelyon/pkg/serp"
	"github.com/oklog/ulid/v2"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

func Work(ctx playwright.BrowserContext, userID ulid.ULID, query string, exclude map[string]bool) {

	g, err := pw.SearchGoogle(query, ctx)
	if err != nil {
		return
	}

	var m *serp.Model
	if m, err = serp.Create(query, pw.Content(g), pw.Screenshot(g)); err != nil {
		return
	}

	Update(userID, query, m.ID)

	defer func() {
		serp.Update(m)
		g.Close()
	}()

	var p playwright.Page
	for _, l := range pw.Locators(g, "[data-dtld]") {
		if domain := pw.Attribute(l, "data-dtld"); exclude[domain] {
			continue
		} else if p, err = pw.NewTab(ctx, l); err != nil || p == nil || p.URL() == "about:blank" {
			continue
		}
		content, screenshot := pw.Content(p), pw.Screenshot(p)
		m.AddSponsored(p.URL(), content)
		serp.Update(m)
		if err = page.SaveScreenshot(p.URL(), m.ID, screenshot); err != nil {
			log.Warn().Err(err).Msg("Failed to save screenshot")
		}
		p.Close()
	}
}
