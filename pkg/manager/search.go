package manager

import (
	"fmt"

	"github.com/nelsw/bytelyon/pkg/client"
	"github.com/nelsw/bytelyon/pkg/db"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/s3"
	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/oklog/ulid/v2"
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

func (j *Job) doSearch() {
	var err error

	var page playwright.Page
	if page, err = client.NewPage(j.ctx); err != nil {
		return
	}
	defer func(page playwright.Page, options ...playwright.PageCloseOptions) {
		_ = page.Close()
	}(page)

	var resp playwright.Response
	if resp, err = client.GoTo(page, "https://www.google.com"); err != nil {
		return
	}

	if client.IsRequestBlocked(resp) || client.IsPageBlocked(page) {
		client.WaitForLoadState(page)
		if client.IsRequestBlocked(resp) || client.IsPageBlocked(page) {
			return
		}
	}

	if err = client.Click(page, googleSearchInputSelectors...); err != nil {
		return
	} else if err = client.Type(page, j.bot.Target); err != nil {
		return
	} else if err = client.Press(page, "Enter"); err != nil {
		return
	} else if err = client.WaitForLoadState(page); err != nil {
		return
	} else if client.IsPageBlocked(page) {
		return
	}

	log.Info().Msgf("Reached Google SERP for query: %s", j.bot.Target)

	result := j.bot.NewBotResult()
	var pages []*model.Page

	pages = append(pages, j.makePage(result, page, 0))

	var locators []playwright.Locator
	if locators, err = page.Locator("[data-dtld]").All(); err != nil {
		log.Warn().Err(err).Msg("Failed to get Locator Count")
		return
	}

	log.Info().Int("locators", len(locators)).Msg("Locators Found")

	var pge *model.Page
	for idx, loc := range locators {

		pge, err = j.handleLocator(result, j.ctx, loc, idx)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to handle locator")
			continue
		}

		pages = append(pages, pge)
	}

	result.Data["pages"] = pages

	if err = db.PutItem(result); err != nil {
		log.Warn().Err(err).Msg("Failed to Save Search Result (DB)")
	}

}

func (j *Job) handleLocator(result *model.BotResult, ctx playwright.BrowserContext, loc playwright.Locator, idx int) (*model.Page, error) {

	var cb = func() error {
		return loc.Click(playwright.LocatorClickOptions{
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

	page, err := ctx.ExpectPage(cb, opt)
	if err != nil {
		log.Warn().Err(err).Msg("Client - Failed to ExpectPage")
		return nil, err
	}
	defer func(page playwright.Page, options ...playwright.PageCloseOptions) {
		_ = page.Close()
	}(page)

	if err = page.BringToFront(); err != nil {
		log.Warn().Err(err).Msg("Client - Failed to BringToFront")
		return nil, err
	}

	if err = client.WaitForLoadState(page, *playwright.LoadStateDomcontentloaded); err != nil {
		log.Warn().Err(err).Msg("Client - Failed to WaitForLoadState")
	}

	return j.makePage(result, page, idx), nil
}

func (j *Job) makePage(result *model.BotResult, page playwright.Page, idx int) *model.Page {

	var p = model.Page{
		URL: page.URL(),
	}

	var err error

	if p.Title, err = page.Title(); err != nil {
		log.Warn().Err(err).Msg("Failed to get page Title")
	}

	var img []byte
	if img, err = page.Screenshot(playwright.PageScreenshotOptions{FullPage: util.Ptr(true)}); err != nil {
		log.Warn().Err(err).Msg("Failed to Screenshot Page")
	} else {
		var s string
		if s, err = s3.PutPublicBotData(j.bot, storagePath(result.ID, "png", idx), img); err == nil {
			p.IMG = s
		}
	}

	var content string
	if content, err = page.Content(); err != nil {
		log.Warn().Err(err).Msg("Failed to get Page Content")
	} else {
		s := storagePath(result.ID, "html", idx)
		if s, err = s3.PutPrivateBotData(j.bot, s, []byte(content)); err == nil {
			p.HTML = s
		}
		if idx == 0 {
			p.SERP = model.MakeSerp(page.URL(), content)
		}
	}

	return &p
}

func storagePath(resultID ulid.ULID, ext string, idx int) string {
	t := "content"
	if ext == "png" {
		t = "screenshot"
	}
	return fmt.Sprintf("%s/%s/%d.%s",
		resultID,
		t,
		idx,
		ext,
	)
}
