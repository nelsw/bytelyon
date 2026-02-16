package search

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/nelsw/bytelyon/internal/client/prowl"
	"github.com/nelsw/bytelyon/internal/db"
	"github.com/nelsw/bytelyon/internal/model"
	"github.com/nelsw/bytelyon/internal/util"
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
}

func New(job *model.Bot) *Worker {
	return &Worker{job}
}

func (w *Worker) Work() {

	err := w.work(true)
	if err == nil {
		return
	}
	log.Err(err).Msgf("Failed to work with headless: %t", true)

	if err = w.work(false); err != nil {
		log.Err(err).Msgf("Failed to work with headless: %t", false)
	}
}

func (w *Worker) work(headless bool) error {
	c, err := prowl.New(headless)
	if err != nil {
		return err
	}
	defer c.Close()

	var google playwright.Page
	if google, err = w.VisitGoogle(c); err != nil {
		return err
	}
	defer google.Close()

	search := &model.Search{Bot: w.Bot}
	if err = db.Create(search); err != nil {
		return err
	}

	if err = w.save(&model.SearchPage{Search: search}, google, c); err != nil {
		log.Warn().Err(err).Msg("Failed to Save Search Page (Google)")
		return err
	}

	var locators []playwright.Locator
	if locators, err = google.Locator(fmt.Sprintf(`[data-dtld]`), playwright.PageLocatorOptions{}).All(); err != nil {
		log.Err(err).Msg("Err finding Locators")
		return err
	} else if len(locators) == 0 {
		log.Warn().Msg("No Target Locators Found")
		return nil
	} else if w.Bot.Ignore()["*"] {
		log.Info().Msg("Ignoring all targets")
		return nil
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
		if _, ok := w.Bot.Ignore()[att]; ok {
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

		if err = w.save(&model.SearchPage{Search: search}, targetPage, c); err != nil {
			log.Warn().Err(err).Msg("Failed to Save Search Page (Target)")
		}
	}
	log.Info().Msg("Finished Search")
	return nil
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

func (w *Worker) save(s *model.SearchPage, page playwright.Page, c *prowl.Client) error {

	s.URL = page.URL()
	if err := db.Create(s); err != nil {
		log.Err(err).Msgf("Failed to Save SearchPage [%s]", s.URL)
		return err
	}

	if title, err := page.Title(); err != nil {
		log.Warn().Err(err).Msg("Failed to get page Title")
	} else {
		s.Title = strings.TrimSpace(title)
	}

	wd, _ := os.Getwd()
	serverPath := fmt.Sprintf("%s/web/bot/search/%d-%d.", wd, s.Search.ID, s.ID)
	staticPath := fmt.Sprintf("/static/bot/search/%d-%d.", s.Search.ID, s.ID)

	if img, err := page.Screenshot(playwright.PageScreenshotOptions{FullPage: util.Ptr(true)}); err != nil {
		log.Warn().Err(err).Msg("Failed to Screenshot SearchPage")
	} else if imgErr := os.WriteFile(serverPath+"png", img, os.ModePerm); imgErr != nil {
		log.Warn().Err(imgErr).Msg("Failed to write SearchPage screenshot")
	} else {
		s.IMG = staticPath + "png"
	}

	if content, err := page.Content(); err != nil {
		log.Warn().Err(err).Msg("Failed to get SearchPage Content")
	} else if htmlErr := os.WriteFile(serverPath+"html", []byte(content), os.ModePerm); htmlErr != nil {
		log.Warn().Err(htmlErr).Msg("Failed to write SearchPage html")
	} else {
		s.HTML = staticPath + "html"
		if strings.Contains(s.URL, "google.com") {
			s.JSON = c.Data(s.URL, content)
		}
	}

	return db.Save(s)
}
