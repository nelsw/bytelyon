package search

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/nelsw/bytelyon/internal/client/prowl"
	"github.com/nelsw/bytelyon/internal/model"
	"github.com/nelsw/bytelyon/internal/service/db"
	"github.com/nelsw/bytelyon/internal/service/s3"
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
	*model.SearchBot
}

func New(job *model.SearchBot) *Worker {
	return &Worker{job}
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

	data := model.SearchBotData{BotID: w.BotID}
	if err = db.Save(&data); err != nil {
		log.Err(err).Msg("Failed to Create Search")
		return
	}

	if err = w.save(data, google, c); err != nil {
		log.Err(err).Msg("Failed to Save Search Page (Google)")
		return
	}

	var locatorCount int
	if locatorCount, err = google.Locator(fmt.Sprintf(`[data-rw]`)).Count(); err != nil {
		log.Err(err).Msg("Err finding Locator Count")
		return
	}

	log.Debug().Int("locators", locatorCount).Msg("Locators Found")

	if w.Bot.Ignore()["*"] {
		log.Info().Msg("Ignoring all targets; Finished Search")
		return
	}

	for i := 0; i < locatorCount; i++ {
		e := w.HandleLocator(c, data, google, i)
		if e != nil {
			err = errors.Join(err, e)
		}
	}

	log.Err(err).Msg("Finished Search")

	w.Bot.UpdatedAt = time.Now()
	if w.Bot.Frequency == 1 {
		w.Bot.Frequency = 0
	}

	err = db.Save(w.Bot)
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

func (w *Worker) HandleLocator(c *prowl.Client, data model.SearchBotData, page playwright.Page, idx int) (err error) {

	l := page.Locator(`[data-rw]`).Nth(idx)
	var att string
	if att, err = l.GetAttribute("data-dtld"); err != nil {
		log.Warn().Err(err).Msg("Failed to get Target Locator Attribute")
		return
	}

	log.Debug().Msgf("Handling Locator [%d] [%s]\n[%s]", idx, att, page.URL())
	if _, ok := w.Bot.Ignore()[att]; ok {
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
	if err = w.save(data, targetPage, c); err != nil {
		log.Warn().Err(err).Msg("Failed to Save Search Page (Target)")
	} else {
		log.Info().Msgf("Saved Search Page [%s]", targetPage.URL())
	}
	err = targetPage.Close()
	return
}

func (w *Worker) save(s model.SearchBotData, page playwright.Page, c *prowl.Client) error {

	p := map[string]any{
		"url": page.URL(),
	}

	if title, err := page.Title(); err != nil {
		log.Warn().Err(err).Msg("Failed to get page Title")
	} else {
		p["title"] = strings.TrimSpace(title)
	}

	ƒ := func(s model.SearchBotData, url, ext string) string {
		return fmt.Sprintf("bots/search/bot/%s/data/%s/%s.%s",
			s.BotID,
			s.DataID,
			base64.URLEncoding.EncodeToString([]byte(url)),
			ext,
		)
	}

	if img, err := page.Screenshot(playwright.PageScreenshotOptions{FullPage: util.Ptr(true)}); err != nil {
		log.Warn().Err(err).Msg("Failed to Screenshot SearchPage")
	} else {
		k := ƒ(s, page.URL(), "png")
		if err = s3.Save(k, img); err != nil {
			log.Warn().Err(err).Msg("Failed to Save Search Page (Screenshot)")
		} else {
			p["screenshot"] = k
		}
	}

	if content, err := page.Content(); err != nil {
		log.Warn().Err(err).Msg("Failed to get SearchPage Content")
	} else {
		k := ƒ(s, page.URL(), "html")
		if err = s3.Save(k, content); err != nil {
			log.Warn().Err(err).Msg("Failed to Save Search Content")
		} else {
			p["content"] = k
			if strings.Contains(page.URL(), "google.com") {
				p["serp"] = c.Data(page.URL(), content)
			}
		}
	}

	return db.Save(&s)
}
