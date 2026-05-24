package search

import (
	"github.com/nelsw/bytelyon/internal/pw"
	"github.com/nelsw/bytelyon/pkg/serp"
	"github.com/nelsw/bytelyon/pkg/snippet"
	"github.com/nelsw/bytelyon/pkg/util/ptr"
	"github.com/oklog/ulid/v2"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

func (m *Model) Run(ctx playwright.BrowserContext, exclude map[string]bool) {

	page, err := pw.SearchGoogle(m.Query, ctx)
	if err != nil {
		return
	}

	defer func() {
		m.Save()
		page.Close()
	}()

	srp := serp.New(page, m.Query)

	srp.Create()
	m.Entries[srp.ID] = append(m.Entries[srp.ID], srp.URL)

	for _, l := range pw.Locators(page, "[data-dtld]") {
		domain := pw.Attribute(l, "data-dtld")
		if exclude[domain] {
			log.Info().Str("domain", domain).Msg("skipping (blacklisted)")
			return
		}
		log.Info().Str("domain", domain).Msg("scraping")

		if snip := m.snippet(ctx, srp.ID, l); snip != nil {
			snip.Create()
			m.Entries[srp.ID] = append(m.Entries[srp.ID], snip.URL)
		}
	}
}

func (m *Model) snippet(ctx playwright.BrowserContext, id ulid.ULID, l playwright.Locator) *snippet.Model {

	var cb = func() error {
		return l.Click(playwright.LocatorClickOptions{
			Force:     ptr.True,
			Modifiers: []playwright.KeyboardModifier{"Meta"},
			Timeout:   ptr.ZeroFloat64,
		})
	}

	var opt = playwright.BrowserContextExpectPageOptions{
		Predicate: func(p playwright.Page) bool { return true },
	}

	page, err := ctx.ExpectPage(cb, opt)
	if err != nil {
		log.Warn().Err(err).Msg("Client - Failed to ExpectPage")
		return nil
	}
	defer page.Close()

	if err = page.BringToFront(); err != nil {
		log.Warn().Err(err).Msg("Client - Failed to BringToFront")
		return nil
	}

	url, content, screenshot := page.URL(), pw.Content(page), pw.Screenshot(page)

	return snippet.New(id, url, content, screenshot)
}
