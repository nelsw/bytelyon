package client

import (
	"regexp"

	. "github.com/nelsw/bytelyon/pkg/util"
	. "github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

var (
	blockedRegex = regexp.MustCompile("(google.com/sorry|captcha|unusual traffic)")
)

// IsPageBlocked determines if we have been blocked from visiting a URL.
func IsPageBlocked(page Page) bool {

	log.Debug().Msg("checking if page is blocked")

	if blockedRegex.MatchString(page.URL()) {
		log.Warn().Msg("page blocked")
		return true
	}

	log.Info().Msg("page reached")
	return false
}

// IsRequestBlocked determines if we have been blocked from requesting a URL.
func IsRequestBlocked(res Response) bool {
	log.Debug().Msg("checking if request is blocked")

	if !res.Ok() {
		log.Warn().Msg("request blocked")
		return true
	}

	log.Info().Msg("response ok")
	return false
}

// Type fills text to type into a focused element.
func Type(page Page, s string) error {
	log.Debug().Str("text", s).Msg("typing")
	err := page.Keyboard().Type(s, KeyboardTypeOptions{
		Delay: Ptr(Between(500.0, 1000.0)),
	})
	if err != nil {
		log.Err(err).Str("text", s).Msg("failed to type")
		return err
	}
	log.Trace().Str("text", s).Msg("typed")
	return err
}

// Press executes a keyboard event on a page.
func Press(page Page, s string) error {
	log.Debug().Str("key", s).Msg("pressing")
	err := page.Keyboard().Press(s, KeyboardPressOptions{
		Delay: Ptr(Between(200, 500.0)),
	})
	if err != nil {
		log.Err(err).Str("key", s).Msg("failed to press")
		return err
	}
	log.Trace().Str("key", s).Msg("pressed")
	return nil
}

// NewPage creates a new page in the browser context.
func NewPage(ctx BrowserContext) (Page, error) {

	log.Debug().Msg("creating new page")

	page, err := ctx.NewPage()
	if err != nil {
		log.Err(err).Msg("failed to create new page")
		return nil, err
	}

	err = page.AddInitScript(Script{Content: Ptr(`() => {
  Object.defineProperty(window.screen, "width", { get: () => 1920 });
  Object.defineProperty(window.screen, "height", { get: () => 1080 });
  Object.defineProperty(window.screen, "colorDepth", { get: () => 24 });
  Object.defineProperty(window.screen, "pixelDepth", { get: () => 24 });
}`)})

	if err != nil {
		log.Err(err).Msg("failed to add init script to page")
		return nil, err
	}

	log.Info().Msg("created new page")

	return page, nil
}

// GoTo returns the main resource response.
func GoTo(page Page, url string) (Response, error) {

	log.Debug().Str("url", url).Msg("go to url")

	res, err := page.Goto(url, PageGotoOptions{
		Timeout:   Ptr(10_000.0),
		WaitUntil: WaitUntilStateDomcontentloaded,
	})

	if err != nil {
		log.Err(err).Str("url", url).Msg("failed to go to url")
		return nil, err
	}

	log.Trace().Str("url", url).Msg("got to url")

	return res, nil
}

// Click the first page element located by the given selectors.
func Click(page Page, selectors ...string) error {

	log.Debug().Strs("selectors", selectors).Msg("clicking")

	var selector string
	var locator Locator
	for _, selector = range selectors {

		if locator = page.Locator(selector); locator == nil {
			continue
		}

		n, err := locator.Count()
		if err != nil {
			log.Err(err).Str("selector", selector).Msg("failed to get locator count")
			return err
		}

		if n == 0 {
			log.Trace().Str("selector", selector).Msg("locator not found")
			continue
		}

		log.Trace().Str("selector", selector).Msg("locator found")
		break
	}

	err := locator.Click(LocatorClickOptions{
		Delay: Ptr(Between(200, 500.0)),
	})

	if err != nil {
		log.Err(err).Str("selector", selector).Msg("failed to click locator")
		return err
	}

	log.Trace().Str("selector", selector).Msg("locator clicked")

	return nil
}

// WaitForLoadState returns nil when the required load state has been reached, or error if an exception occurred.
func WaitForLoadState(page Page, ls ...LoadState) error {

	s := LoadStateNetworkidle
	if len(ls) > 0 {
		s = &ls[0]
	}

	log.Debug().Any("state", s).Msg("waiting for load state")

	err := page.WaitForLoadState(PageWaitForLoadStateOptions{
		State: s,
	})
	if err != nil {
		log.Err(err).Any("state", s).Msg("failed to reach load state")
		return err
	}

	log.Trace().Msg("load state reached")
	return nil
}

// Content returns the page content or an empty string if the page has failed to load.
func Content(page Page) string {
	s, err := page.Content()
	if err != nil {
		log.Err(err).Msg("failed to get page content")
		return ""
	}
	return s
}

// Screenshot returns the screenshot of the page as a byte array or an empty byte array if the page has failed to load.
func Screenshot(page Page, opts ...PageScreenshotOptions) []byte {
	opts = append(opts, PageScreenshotOptions{FullPage: Ptr(true)})
	b, err := page.Screenshot()
	if err != nil {
		log.Err(err).Msg("failed to get page screenshot")
		return nil
	}
	return b
}

// Title returns the page title or an empty string if the page has failed to load.
func Title(page Page) string {
	s, err := page.Title()
	if err != nil {
		log.Err(err).Msg("failed to get page title")
		return ""
	}
	return s
}
