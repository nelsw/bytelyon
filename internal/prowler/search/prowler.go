package search

import (
	"github.com/nelsw/bytelyon/internal/pw"
	"github.com/nelsw/bytelyon/pkg/em"
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
}

func New(bot *model.Bot, ctx playwright.BrowserContext) *Prowler {
	return &Prowler{
		ctx:    ctx,
		Search: entity.NewSearch(bot),
	}
}

func (p *Prowler) Prowl() {
	defer em.PutSearch(p.Search)
	p.prowlSearchPage()
}

func (p *Prowler) prowlSearchPage() {
	searchPage, err := pw.SearchGoogle(p.Target, p.ctx)
	if err != nil {
		return
	}
	defer searchPage.Close()

	p.Add(entity.NewPage(searchPage))
	for _, l := range pw.Locators(searchPage, "[data-dtld]") {
		// todo - blacklist
		domain := pw.Attribute(l, "data-dtld")
		log.Info().Str("data-dtld", domain).Msg("Prowl")
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

	page, err := p.ctx.ExpectPage(cb, opt)
	if err != nil {
		log.Warn().Err(err).Msg("Client - Failed to ExpectPage")
		return
	}
	defer page.Close()

	if err = page.BringToFront(); err != nil {
		log.Warn().Err(err).Msg("Client - Failed to BringToFront")
		return
	}

	p.Add(entity.NewPage(page))
}
