package search

import (
	"github.com/nelsw/bytelyon/internal/pw"
	"github.com/nelsw/bytelyon/pkg/entity"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/util/ptr"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

type Prowler struct {

	// ctx is the context of the browser, which is used to run the browser and the page
	ctx playwright.BrowserContext

	*entity.Search

	blackMap map[string]bool
}

func New(bot *model.Bot, ctx playwright.BrowserContext) *Prowler {
	return &Prowler{
		ctx:      ctx,
		Search:   new(entity.Search).From(bot.UserID, bot.Target),
		blackMap: bot.BlackMap(),
	}
}

func (p *Prowler) Prowl() {
	p.prowlSearchPage()
}

func (p *Prowler) prowlSearchPage() {
	searchPage, err := pw.SearchGoogle(p.Query, p.ctx)
	if err != nil {
		return
	}
	defer searchPage.Close()

	page := entity.NewPage(searchPage)
	page.Save()

	p.Serp = page.SERP
	p.Save()

	for _, l := range pw.Locators(searchPage, "[data-dtld]") {
		domain := pw.Attribute(l, "data-dtld")
		if p.blackMap[domain] {
			log.Info().Str("domain", domain).Msg("skipping (blacklisted)")
			return
		}
		log.Info().Str("domain", domain).Msg("scraping")
		p.prowlResultPages(l)
	}
}

func (p *Prowler) prowlResultPages(l playwright.Locator) {

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

	resultPage, err := p.ctx.ExpectPage(cb, opt)
	if err != nil {
		log.Warn().Err(err).Msg("Client - Failed to ExpectPage")
		return
	}
	defer resultPage.Close()

	if err = resultPage.BringToFront(); err != nil {
		log.Warn().Err(err).Msg("Client - Failed to BringToFront")
		return
	}

	page := entity.NewPage(resultPage)
	page.Save()

	p.Snippets = append(p.Snippets, page.MakeSnippet())
	p.Save()
}
