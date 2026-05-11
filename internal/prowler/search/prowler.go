package search

import (
	"github.com/nelsw/bytelyon/internal/pw"
	"github.com/nelsw/bytelyon/pkg/em"
	"github.com/nelsw/bytelyon/pkg/entity"
	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/oklog/ulid/v2"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

type Prowler struct {

	// ctx is the context of the browser, which is used to run the browser and the page
	ctx playwright.BrowserContext

	e *entity.Search
}

func New(q string, ctx playwright.BrowserContext) *Prowler {
	return &Prowler{
		ctx: ctx,
		e:   entity.NewSearch(q),
	}
}

func (p *Prowler) Prowl(userID ulid.ULID) {
	defer func() {
		em.SaveSearch(userID, p.e)
	}()
	p.prowlSearchPage()
}

func (p *Prowler) prowlSearchPage() {
	searchPage, err := pw.SearchGoogle(p.e.Query, p.ctx)
	if err != nil {
		return
	}
	defer searchPage.Close()

	p.e.AddPage(entity.NewPage(searchPage))
	// todo - blacklist
	for _, l := range pw.Locators(searchPage, "[data-dtld]") {
		p.prowlResultPages(l)
	}
}

func (p *Prowler) prowlResultPages(l playwright.Locator) {

	var cb = func() error {
		return l.Click(playwright.LocatorClickOptions{
			Force: util.Ptr(true),
			Modifiers: []playwright.KeyboardModifier{
				*playwright.KeyboardModifierMeta,
			},
			Timeout: util.Ptr(0.0),
		})
	}

	var opt = playwright.BrowserContextExpectPageOptions{
		Predicate: func(p playwright.Page) bool {
			return true
		},
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

	p.e.AddPage(entity.NewPage(resultPage))
}
