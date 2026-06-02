package search

import (
	"github.com/nelsw/bytelyon/pkg/page"
	"github.com/nelsw/bytelyon/pkg/pw"
	"github.com/nelsw/bytelyon/pkg/serp"
	"github.com/oklog/ulid/v2"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

func Work(ctx playwright.BrowserContext, userID ulid.ULID, query string, exclude map[string]bool) {

	g, err := pw.SearchGoogle(query, ctx)
	if err != nil {
		log.Err(err).Msgf("work search %s", query)
		return
	}

	var m *serp.Model
	if m, err = serp.Create(query, pw.Content(g), pw.Screenshot(g)); err != nil {
		log.Err(err).Msgf("work search %s", query)
		return
	}

	if err = Update(userID, query, m.ID); err != nil {
		log.Err(err).Msg("Failed to save search object")
		return
	}

	defer func() {
		if closeErr := g.Close(); closeErr != nil {
			log.Warn().Err(closeErr).Msg("Failed to close page")
		}
	}()

	var p playwright.Page
	for _, l := range pw.Locators(g, "[data-dtld]") {
		if domain := pw.Attribute(l, "data-dtld"); exclude[domain] {
			continue
		}

		if p, err = pw.NewTab(ctx, l); err != nil || p == nil || p.URL() == "about:blank" {
			if p != nil {
				if err = p.Close(); err != nil {
					log.Warn().Err(err).Msg("Failed to close page")
				}
			}
			continue
		}
		content, screenshot := pw.Content(p), pw.Screenshot(p)
		m.AddSponsored(p.URL(), content)
		if err = serp.Update(query, m); err != nil {
			log.Warn().Err(err).Msg("Failed to save serp object")
		}
		if err = page.SaveScreenshot(p.URL(), m.ID, screenshot); err != nil {
			log.Warn().Err(err).Msg("Failed to save screenshot")
		}
		if err = p.Close(); err != nil {
			log.Warn().Err(err).Msg("Failed to close page")
		}
	}
}
