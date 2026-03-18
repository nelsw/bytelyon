package search

import (
	"errors"
	"fmt"

	"github.com/nelsw/bytelyon/internal/client/prowl"
	"github.com/nelsw/bytelyon/internal/service/s3"
	"github.com/nelsw/bytelyon/internal/util"
	"github.com/nelsw/bytelyon/pkg/db"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

var googleSearchInputSelectors = []string{
	"input[name='q']",
	"input[title='Search']",
	"input[aria-label='Search']",
	"textarea[title='Search']",
	"textarea[name='q']",
	"textarea[aria-label='Search']",
	"textarea",
}

type Worker struct {
	*model.Bot
	Ignore map[string]bool
	pages  []*model.Page
}

func New(b *model.Bot) *Worker {

	w := &Worker{
		b,
		make(map[string]bool),
		[]*model.Page{},
	}

	if len(b.BlackList) > 0 {
		for _, s := range b.BlackList {
			w.Ignore[s] = true
		}
	}

	return w
}

func (w *Worker) Work() {

	c, err := prowl.New(w.Headless)
	if err != nil {
		log.Err(err).Msg("Failed to create Prowl Client")
		return
	}
	defer c.Close()

	var google playwright.Page
	if google, err = w.VisitGoogle(c); err != nil {
		log.Err(err).Msg("Failed to Visit Google")
		return
	}

	if err = w.save(google); err != nil {
		log.Err(err).Msg("Failed to Save Search Page (Google)")
		return
	}

	var locatorCount int
	if locatorCount, err = google.Locator(fmt.Sprintf(`[data-rw]`)).Count(); err != nil {
		log.Err(err).Msg("Err finding Locator Count")
		return
	}

	log.Info().Int("locators", locatorCount).Msg("Locators Found")

	if _, ok := w.Ignore["*"]; ok {
		log.Info().Msg("Ignoring all targets; Finished Search")
		return
	}

	for i := 0; i < locatorCount; i++ {
		if e := w.HandleLocator(c, google, i); e != nil {
			err = errors.Join(err, e)
		}
	}

	log.Err(err).Msg("Finished Search")
}

func (w *Worker) VisitGoogle(c *prowl.Client) (page playwright.Page, err error) {

	var res playwright.Response

	if page, err = c.NewPage(); err != nil {
		return
	} else if res, err = c.GoTo(page, "https://www.google.com"); err != nil {
		return
	} else if err = c.IsBlocked(page, res); err != nil {
		return
	} else if err = c.Click(page, googleSearchInputSelectors...); err != nil {
		return
	} else if err = c.Type(page, w.Target); err != nil {
		return
	} else if err = c.Press(page, "Enter"); err != nil {
		return
	} else if err = c.WaitForLoadState(page); err != nil {
		return
	} else if err = c.IsBlocked(page); err != nil {
		return
	}

	log.Info().Msgf("Visited Google with query: %s", w.Target)

	c.SetState()

	return
}

func (w *Worker) HandleLocator(c *prowl.Client, page playwright.Page, idx int) (err error) {

	l := page.Locator(`[data-rw]`).Nth(idx)
	var att string
	if att, err = l.GetAttribute("data-dtld"); err != nil {
		log.Warn().Err(err).Msg("Failed to get Target Locator Attribute")
		return
	}

	log.Debug().Msgf("Handling Locator [%d] [%s]\n[%s]", idx, att, page.URL())
	if _, ok := w.Ignore[att]; ok {
		return
	}

	var targetPage playwright.Page
	if targetPage, err = c.BrowserContext.ExpectPage(func() error {
		return l.Click(playwright.LocatorClickOptions{
			Force: util.Ptr(true),
			Modifiers: []playwright.KeyboardModifier{
				*playwright.KeyboardModifierMeta,
			},
			Timeout: util.Ptr(0.0),
		})
	}, playwright.BrowserContextExpectPageOptions{
		Predicate: func(p playwright.Page) bool {
			return true
		},
	}); err != nil {
		log.Warn().Err(err).Msg("Client - Failed to ExpectPage")
		return err
	} else if err = page.BringToFront(); err != nil {
		log.Warn().Err(err).Msg("Client - Failed to BringToFront")
		return err
	}

	if err = c.WaitForLoadState(targetPage, *playwright.LoadStateDomcontentloaded); err != nil {
		log.Warn().Err(err).Msg("Client - Failed to WaitForLoadState")
	}

	log.Debug().Int("pages", len(c.BrowserContext.Pages())).Msg("Pages")
	if err = w.save(targetPage); err != nil {
		log.Warn().Err(err).Msg("Failed to Save Search Page (Target)")
	} else {
		log.Info().Msgf("Saved Search Page [%s]", targetPage.URL())
	}
	err = targetPage.Close()
	return
}

func (w *Worker) save(page playwright.Page) (err error) {

	p := model.Page{
		URL: page.URL(),
	}

	if p.Title, err = page.Title(); err != nil {
		log.Warn().Err(err).Msg("Failed to get page Title")
	}

	var img []byte
	if img, err = page.Screenshot(playwright.PageScreenshotOptions{FullPage: util.Ptr(true)}); err != nil {
		log.Warn().Err(err).Msg("Failed to Screenshot Page")
	}

	if len(img) > 0 {
		p.IMG = w.StoragePath(len(w.pages), "png")
		if err = s3.Save(p.IMG, img); err != nil {
			log.Warn().Err(err).Msg("Failed to Save Search Page (Screenshot)")
		}
	}

	var content string
	if content, err = page.Content(); err != nil {
		log.Warn().Err(err).Msg("Failed to get Page Content")
	}

	if len(content) > 0 {
		p.HTML = w.StoragePath(len(w.pages), "html")
		if err = s3.Save(p.HTML, []byte(content)); err != nil {
			log.Warn().Err(err).Msg("Failed to Save Search Page (HTML)")
		}
		p.SERP = model.MakeSerp(page.URL(), content)
	}

	w.pages = append(w.pages, &p)
	if err = db.PutItem(w.NewBotResult("pages", w.pages)); err != nil {
		log.Error().Err(err).Msg("Failed to Save Search Page (DB)")
	}
	return
}
