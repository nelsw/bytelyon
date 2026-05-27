package search

import (
	"github.com/nelsw/bytelyon/internal/pw"
	"github.com/nelsw/bytelyon/pkg/serp"
	"github.com/nelsw/bytelyon/pkg/snippet"
	"github.com/oklog/ulid/v2"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

func Work(ctx playwright.BrowserContext, userID ulid.ULID, query string, exclude map[string]bool) {

	page, err := pw.SearchGoogle(query, ctx)
	if err != nil {
		return
	}

	srp := serp.New(query, pw.Content(page), pw.Screenshot(page))
	if err = srp.Save(); err != nil {
		log.Warn().Err(err).Msg("Failed to save serp")
		return
	}

	m := map[ulid.ULID][]string{
		srp.ID: {srp.URL},
	}

	defer func() {
		Save(userID, query, m)
		page.Close()
	}()

	var p playwright.Page
	for _, l := range pw.Locators(page, "[data-dtld]") {

		if domain := pw.Attribute(l, "data-dtld"); exclude[domain] {
			continue
		} else if p, err = pw.NewTab(ctx, l); err != nil || p == nil {
			continue
		}

		snip := snippet.New(srp.ID, p.URL(), pw.Content(p), pw.Screenshot(p))
		if err = snip.Save(); err != nil {
			log.Warn().Err(err).Msg("Failed to save snippet")
		}
		m[srp.ID] = append(m[srp.ID], snip.URL)
		p.Close()
	}
}
