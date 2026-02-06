package search

import (
	"fmt"
	"time"

	"github.com/nelsw/bytelyon/internal/client/prowl"
	"github.com/nelsw/bytelyon/internal/model"
	"github.com/nelsw/bytelyon/internal/util"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

var (
	googleSearchInputSelectors = []string{
		"input[name='q']",
		"input[title='Search']",
		"input[aria-label='Search']",
		"textarea[title='Search']",
		"textarea[name='q']",
		"textarea[aria-label='Search']",
		"textarea",
	}
)

type Worker struct {
	*model.Job
}

func New(job *model.Job) *Worker {
	return &Worker{job}
}

func (w *Worker) Work() *model.Search {
	arr, err := w.work(true)
	if err != nil {
		log.Err(err).Msgf("Failed to work with headless: %t", true)
		if arr, err = w.work(false); err != nil {
			log.Err(err).Msgf("Failed to work with headless: %t", false)
		}
	}
	return arr
}

func (w *Worker) work(headless bool) (*model.Search, error) {
	c, err := prowl.New(headless)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	var google playwright.Page
	if google, err = w.VisitGoogle(c); err != nil {
		return nil, err
	}
	defer google.Close()

	a := &model.Search{
		Pages: []*model.SearchPage{w.toModel(google)},
	}

	var locators []playwright.Locator
	if locators, err = google.Locator(fmt.Sprintf(`[data-dtld]`), playwright.PageLocatorOptions{}).All(); err != nil {
		log.Warn().Err(err).Msg("No Target Locators Found")
		return a, nil
	} else if len(locators) == 0 {
		log.Warn().Msg("No Target Locators Found")
		return a, nil
	} else if w.Job.Ignore()["*"] {
		log.Info().Msg("Ignoring all targets")
		return a, nil
	}

	var att string
	for _, l := range locators {

		if att, err = l.GetAttribute("data-dtld", playwright.LocatorGetAttributeOptions{
			Timeout: util.Ptr(5_000.0),
		}); err != nil {
			log.Warn().Err(err).Msg("Failed to get Target Locator Attribute")
			continue
		}

		log.Debug().Str("found", att).Msg("Locator")
		if _, ok := w.Job.Ignore()[att]; ok {
			continue
		}

		log.Info().Msgf("Target Found [%s]", att)

		if err = l.Click(playwright.LocatorClickOptions{Timeout: util.Ptr(5_000.0)}); err != nil {
			log.Warn().Err(err).Msg("Failed to Click Target Locator")
			continue
		}

		targetPage, pageErr := c.NewPage(func() error { return l.Click() })
		if pageErr != nil {
			log.Warn().Err(pageErr).Msg("Failed to Click Target")
			continue
		}

		a.Pages = append(a.Pages, w.toModel(targetPage))
	}

	return a, nil
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
	}

	if err = c.IsBlocked(page); err != nil && *c.Headless {
		return
	} else if !*c.Headless {
		time.Sleep(time.Second * 15)
		if err = c.IsBlocked(page); err != nil {
			return
		}
	}

	log.Info().Msgf("Visited Google with query: %s", w.Target)

	c.SetState()

	return
}

func (w *Worker) toModel(page playwright.Page) *model.SearchPage {
	var err error
	var img []byte
	if img, err = page.Screenshot(playwright.PageScreenshotOptions{FullPage: util.Ptr(true)}); err != nil {
		log.Warn().Err(err).Msg("PW - Failed to Screenshot SearchPage")
	}

	var content string
	if content, err = page.Content(); err != nil {
		log.Warn().Err(err).Msg("PW - Failed to get SearchPage Content")
	}

	var title string
	if title, err = page.Title(); err != nil {
		log.Warn().Err(err).Msg("PW - Failed to get SearchPage Title")
	}
	return &model.SearchPage{
		URL:   page.URL(),
		Title: title,
		HTML:  content,
		IMG:   img,
	}
}
