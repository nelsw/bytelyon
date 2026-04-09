package pw

import (
	"time"

	"github.com/nelsw/bytelyon/pkg/client"
	"github.com/nelsw/bytelyon/pkg/logs"
	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

var (
	Client *playwright.Playwright
)

func Init() {
	log.Debug().Msg("installing playwright drivers")
	if err := playwright.Install(&playwright.RunOptions{Logger: logs.NewSlog()}); err != nil {
		log.Panic().Err(err).Msg("failed to install playwright")
	}

	log.Debug().Msg("starting a playwright instance")
	var err error
	if Client, err = playwright.Run(); err != nil {
		log.Panic().Err(err).Msg("failed to start playwright")
	}

	log.Info().Msg("playwright initialized successfully")
}

func NewBrowser(headless bool) (playwright.Browser, error) {
	return client.NewBrowser(Client, headless)
}

func NewBrowsers() (headless playwright.Browser, headed playwright.Browser, err error) {
	if headless, err = client.NewBrowser(Client, true); err != nil {
		return
	}
	if headed, err = client.NewBrowser(Client, false); err != nil {
		headless.Close()
	}
	return
}

func FetchPageData(url string, ctx playwright.BrowserContext) (title string, source string, image []byte, err error) {

	var page playwright.Page

	if page, err = client.NewPage(ctx); err != nil {
		log.Err(err).Msg("failed to create new page for data fetch")
		return
	}
	defer func(page playwright.Page) {
		_ = page.Close()
	}(page)

	var resp playwright.Response
	if resp, err = client.GoTo(page, url); err != nil {
		log.Err(err).Str("url", url).Msg("failed to go to page url for data fetch")
		return
	}

	log.Debug().
		Str("url", url).
		Int("resp", resp.Status()).
		Msg("got to page url for data fetch")

	if client.IsRequestBlocked(resp) {
		log.Warn().Str("url", url).Msg("request is blocked for data fetch; returning")
		return
	}

	if client.IsPageBlocked(page) {
		log.Warn().Str("url", url).Msg("page is blocked for data fetch ... waiting for user interaction")
		time.Sleep(time.Minute)
		if err = client.WaitForLoadState(page); err != nil {
			log.Warn().Err(err).Msg("failed to wait for load state")
		}
	}

	if client.IsPageBlocked(page) {
		log.Warn().Str("url", url).Msg("page is still blocked for data fetch ... returning")
		return
	}

	log.Info().
		Str("url", url).
		Msg("fetched page data")

	if title, err = page.Title(); err != nil {
		log.Warn().Err(err).Msg("failed to get title from page")
	}

	if source, err = page.Content(); err != nil {
		log.Warn().Err(err).Msg("failed to get content from page")
	}

	if image, err = page.Screenshot(playwright.PageScreenshotOptions{FullPage: util.Ptr(true)}); err != nil {
		log.Warn().Err(err).Msg("failed to get screenshot from page")
	}

	return
}
