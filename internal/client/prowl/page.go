package prowl

import (
	"errors"

	. "github.com/nelsw/bytelyon/internal/util"

	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

const (
	pageScriptContent = `() => {
  Object.defineProperty(window.screen, "width", { get: () => 1920 });
  Object.defineProperty(window.screen, "height", { get: () => 1080 });
  Object.defineProperty(window.screen, "colorDepth", { get: () => 24 });
  Object.defineProperty(window.screen, "pixelDepth", { get: () => 24 });
}`
)

func (c *Client) NewPage(ff ...func() error) (playwright.Page, error) {

	if len(ff) > 0 {
		page, err := c.BrowserContext.ExpectPage(ff[0])
		if err != nil {
			log.Warn().Err(err).Msg("Client - Failed to ExpectPage")
		}
		page.BringToFront()
		return page, err
	}

	page, err := c.BrowserContext.NewPage()
	if err != nil {
		log.Warn().Err(err).Msg("Client - Failed to NewPage")
	} else if err = page.AddInitScript(playwright.Script{Content: Ptr(pageScriptContent)}); err != nil {
		log.Warn().Err(err).Msg("Client - Failed to AddInitScript")
	}

	return page, err
}

func (c *Client) GoTo(page playwright.Page, url string) (playwright.Response, error) {

	res, err := page.Goto(url, playwright.PageGotoOptions{
		Timeout:   Ptr(10_000.0),
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})

	if err == nil && !res.Ok() {
		err = errors.New(res.StatusText())
	}

	log.Err(err).Str("url", url).Msg("Client - GoTo")

	return res, err
}

func (c *Client) Click(page playwright.Page, selectors ...string) (err error) {

	var locator playwright.Locator
	for _, selector := range selectors {

		if locator = page.Locator(selector); locator == nil {
			continue
		}

		var n int
		if n, err = locator.Count(); n == 0 {
			continue
		}

		if err = locator.Click(playwright.LocatorClickOptions{Delay: Ptr(Between(200, 500.0))}); err == nil {
			log.Info().Str("selector", selector).Msg("Client - Click")
			return nil
		}

		log.Warn().Err(err).Str("selector", selector).Msg("Client - Failed to Click")
	}

	return err
}

func (c *Client) WaitForLoadState(page playwright.Page, ls ...playwright.LoadState) error {
	s := playwright.LoadStateNetworkidle
	if len(ls) > 0 {
		s = &ls[0]
	}
	err := page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State:   s,
		Timeout: Ptr(60_000.0),
	})
	if err != nil {
		log.Err(err).Msg("Client - WaitForLoadState")
	}
	return err
}
