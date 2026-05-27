package search

import (
	"github.com/nelsw/bytelyon/internal/pw"
	"github.com/nelsw/bytelyon/pkg/document"
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

	srp := serp.New(query, pw.Content(g), pw.Screenshot(g))
	if err = page.SaveObject(srp.URL, srp.ID, srp); err != nil {
		log.Warn().Err(err).Msg("Failed to save serp object")
		return
	}
	if err = page.SaveContent(srp.URL, srp.ID, srp.Content); err != nil {
		log.Warn().Err(err).Msg("Failed to save serp content")
		return
	}
	if err = page.SaveScreenshot(srp.URL, srp.ID, srp.Screenshot); err != nil {
		log.Warn().Err(err).Msg("Failed to save serp screenshot")
		return
	}

	defer func() {
		Save(userID, query, map[ulid.ULID]string{
			srp.ID: srp.URL,
		})
		g.Close()
	}()

	var p playwright.Page
	for _, l := range pw.Locators(g, "[data-dtld]") {

		if domain := pw.Attribute(l, "data-dtld"); exclude[domain] {
			continue
		} else if p, err = pw.NewTab(ctx, l); err != nil || p == nil {
			continue
		}

		if err = page.SaveScreenshot(p.URL(), srp.ID, pw.Screenshot(p)); err != nil {
			log.Warn().Err(err).Msg("Failed to save screenshot")
		}
		doc := document.New(pw.Content(p))
		srp.AddSponsored(map[string]any{
			"link":    p.URL(),
			"title":   doc.Meta.Title(),
			"snippet": doc.Meta.Description(),
			"source":  doc.Meta.Source(),
		})
		if err = page.SaveObject(srp.URL, srp.ID, srp); err != nil {
			log.Warn().Err(err).Msg("Failed to save serp object")
		}
		p.Close()
	}
}
