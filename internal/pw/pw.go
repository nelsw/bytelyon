package pw

import (
	"errors"
	"fmt"
	"maps"
	"math/rand"
	"regexp"
	"slices"
	"strings"

	"github.com/nelsw/bytelyon/pkg/logs"
	"github.com/nelsw/bytelyon/pkg/model"
	. "github.com/nelsw/bytelyon/pkg/util"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

var (
	blockedRegex = regexp.MustCompile(`(google.com/sorry|captcha|unusual traffic)`)
	hrefSchemes  = regexp.MustCompile(`^(mailto|tel|sms|fax|callto|geo):.*`)

	googleSearchInputSelectors = []string{
		"input[name='q']",
		"input[title='Search']",
		"input[aria-label='Search']",
		"textarea[title='Search']",
		"textarea[name='q']",
		"textarea[aria-label='Search']",
		"textarea",
	}
	installed = false
)

// Install gets drivers from the document and installs them to the machine running ByteLyon
func Install() {
	if installed {
		return
	}
	if err := playwright.Install(&playwright.RunOptions{Logger: logs.NewSlog()}); err != nil {
		panic(err)
	}
	installed = true
}

// Run creates a new Playwright instance
func Run() *playwright.Playwright {

	if !installed {
		Install()
	}

	pwc, err := playwright.Run()
	if err != nil {
		panic(err)
	}
	return pwc
}

// NewBrowser creates a new Browser instance
func NewBrowser(c *playwright.Playwright, headless bool) (playwright.Browser, error) {
	return c.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: &headless,
		Timeout:  Ptr(2 * 60_000.0),
		Args: []string{
			"--disable-accelerated-2d-canvas",
			"--disable-background-networking",
			"--disable-background-timer-throttling",
			"--disable-backgrounding-occluded-windows",
			"--disable-blink-features=AutomationControlled",
			"--disable-breakpad",
			"--disable-component-extensions-with-background-pages",
			"--disable-dev-shm-usage",
			"--disable-extensions",
			"--disable-features=IsolateOrigins,site-per-process",
			"--disable-features=TranslateUI",
			"--disable-gpu",
			"--disable-ipc-flooding-protection",
			"--disable-renderer-backgrounding",
			"--disable-setuid-sandbox",
			"--disable-site-isolation-trials",
			"--disable-document-security",
			"--enable-features=NetworkService,NetworkServiceInProcess",
			"--force-color-profile=srgb",
			"--hide-scrollbars",
			"--metrics-recording-only",
			"--mute-audio",
			"--no-first-run",
			"--no-sandbox",
			"--no-zygote",
		},
		IgnoreDefaultArgs: []string{
			"--enable-automation",
		},
	})
}

// NewBrowserContext creates a new BrowserContext instance
func NewBrowserContext(bro playwright.Browser, state *playwright.OptionalStorageState) (playwright.BrowserContext, error) {

	userAgent := func() *string {

		agents := []string{
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.0.0 Safari/537.36",
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.5112.79 Safari/537.36",
			"Mozilla/5.0 (Windows 7 Enterprise; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.6099.71 Safari/537.36",
			"Mozilla/5.0 (Windows NT 10.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.0.0 Safari/537.36",
			"Mozilla/5.0 (Windows NT 10.0; WOW64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.5756.197 Safari/537.36",
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.3713.147 Safari/537.36",
			"Mozilla/5.0 (Windows NT 11.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.6998.166 Safari/537.36",
			"Mozilla/5.0 (Windows Server 2012 R2 Standard; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.5975.80 Safari/537.36",
			"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.5060.53 Safari/537.36",
			"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.5672 Safari/537.36",
			"Mozilla/5.0 (X11; Linux x86_64; CentOS Ubuntu 19.04) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.5957.0 Safari/537.36",
		}

		return Ptr(agents[rand.Intn(len(agents))])
	}

	ctx, err := bro.NewContext(playwright.BrowserNewContextOptions{
		AcceptDownloads:   Ptr(true),
		ColorScheme:       playwright.ColorSchemeDark,
		ForcedColors:      playwright.ForcedColorsNone,
		HasTouch:          Ptr(false),
		IsMobile:          Ptr(false),
		JavaScriptEnabled: Ptr(true),
		Locale:            Ptr("en-US"),
		Permissions:       []string{"geolocation", "notifications"},
		ReducedMotion:     playwright.ReducedMotionNoPreference,
		TimezoneId:        Ptr("America/New_York"),
		UserAgent:         userAgent(),
		StorageState:      state,
	})

	if err != nil {
		return nil, err
	}

	err = ctx.AddInitScript(playwright.Script{Content: Ptr(`() => {
  // navigator
  Object.defineProperty(navigator, "webdriver", { get: () => false });
  Object.defineProperty(navigator, "plugins", {
	get: () => [1, 2, 3, 4, 5],
  });
  Object.defineProperty(navigator, "languages", {
	get: () => ["en-US", "en", "zh-CN"],
  });

  // window
  window.chrome = {
	runtime: {},
	loadTimes: function () {},
	csi: function () {},
	app: {},
  };

  // WebGL
  if (typeof WebGLRenderingContext !== "undefined") {
	const getParameter = WebGLRenderingContext.prototype.getParameter;
	WebGLRenderingContext.prototype.getParameter = function (
	  parameter: number
	) {
	  // UNMASKED_VENDOR_WEBGL / UNMASKED_RENDERER_WEBGL
	  if (parameter === 37445) {
		return "Intel Inc.";
	  }
	  if (parameter === 37446) {
		return "Intel Iris OpenGL Engine";
	  }
	  return getParameter.call(this, parameter);
	};
  }
}`)})

	if err != nil {
		return nil, err
	}

	ctx.SetDefaultTimeout(60_000)

	return ctx, nil
}

// IsPageBlocked determines if we have been blocked from visiting a URL.
func IsPageBlocked(page playwright.Page) bool {
	return blockedRegex.MatchString(page.URL())
}

// IsRequestBlocked determines if we have been blocked from requesting a URL.
func IsRequestBlocked(res playwright.Response) bool {
	return !res.Ok()
}

// Type fills text to type into a focused element.
func Type(page playwright.Page, s string) error {
	return page.Keyboard().Type(s, playwright.KeyboardTypeOptions{
		Delay: Ptr(Between(500.0, 1000.0)),
	})
}

// Press executes a keyboard event on a document.
func Press(page playwright.Page, s string) error {
	return page.Keyboard().Press(s, playwright.KeyboardPressOptions{
		Delay: Ptr(Between(200, 500.0)),
	})
}

// NewPage creates a new document in the browser context.
func NewPage(ctx playwright.BrowserContext) (page playwright.Page, err error) {
	if page, err = ctx.NewPage(); err == nil {
		err = page.AddInitScript(playwright.Script{Content: Ptr(`() => {
  Object.defineProperty(window.screen, "width", { get: () => 1920 });
  Object.defineProperty(window.screen, "height", { get: () => 1080 });
  Object.defineProperty(window.screen, "colorDepth", { get: () => 24 });
  Object.defineProperty(window.screen, "pixelDepth", { get: () => 24 });
}`)})
	}
	return
}

// GoTo returns the main resource response.
func GoTo(page playwright.Page, url string) (playwright.Response, error) {
	return page.Goto(url, playwright.PageGotoOptions{
		Timeout:   Ptr(10_000.0),
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
}

// Visit navigates to the given URL, waits for the document to load,
// and scrolls to the bottom of the page to ensure all content is visible.
func Visit(page playwright.Page, url string) error {

	if res, err := GoTo(page, url); err != nil {
		return err
	} else if !res.Ok() {
		return fmt.Errorf("failed to visit %s: [%d] %s", url, res.Status(), res.StatusText())
	} else if err = WaitForLoadState(page, "networkidle"); err != nil {
		return err
	}

	_, err := page.Evaluate(`async () => {
  await new Promise((resolve) => {
    let totalHeight = 0;
    let distance = 100;
    let timer = setInterval(() => {
      let scrollHeight = document.body.scrollHeight;
      window.scrollBy(0, distance);
      totalHeight += distance;
      if (totalHeight >= scrollHeight) {
		window.scrollTo(0, 0);
        clearInterval(timer);
        resolve();
      }
    }, 100);
  });
}`)

	return err
}

// Click the first document element located by the given selectors.
func Click(page playwright.Page, selectors ...string) (err error) {

	var count int
	var selector string
	var locator playwright.Locator
	for _, selector = range selectors {

		if locator = page.Locator(selector); locator == nil {
			continue
		}

		if count, err = locator.Count(); err != nil || count == 0 {
			continue
		}

		if err = locator.Click(playwright.LocatorClickOptions{Delay: Ptr(Between(200, 500.0))}); err != nil {
			continue
		}

		break
	}
	return
}

// WaitForLoadState returns nil when the required load state has been reached, or error if an exception occurred.
func WaitForLoadState(page playwright.Page, ls ...playwright.LoadState) error {
	s := playwright.LoadStateNetworkidle
	if len(ls) > 0 {
		s = &ls[0]
	}
	return page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: s,
	})
}

// Content returns the document content or an empty string if the document has failed to load.
func Content(page playwright.Page) string {
	s, err := page.Content()
	if err != nil {
		log.Err(err).Msg("failed to get document content")
		return ""
	}
	return s
}

// Screenshot returns the screenshot of the document as a byte array or an empty byte array if the document has failed to load.
func Screenshot(page playwright.Page) []byte {
	/*
		await page.evaluate(async () => {
		  await new Promise((resolve) => {
		    let totalHeight = 0;
		    let distance = 100;
		    let timer = setInterval(() => {
		      let scrollHeight = document.body.scrollHeight;
		      window.scrollBy(0, distance);
		      totalHeight += distance;
		      if (totalHeight >= scrollHeight) {
		        clearInterval(timer);
		        resolve();
		      }
		    }, 100);
		  });
		});
	*/
	b, err := page.Screenshot(playwright.PageScreenshotOptions{FullPage: Ptr(true)})
	if err != nil {
		log.Err(err).Msg("failed to get document screenshot")
		return nil
	}
	return b
}

// Title returns the document title or an empty string if the document has failed to load.
func Title(page playwright.Page) string {
	s, err := page.Title()
	if err != nil {
		log.Err(err).Msg("failed to get document title")
		return ""
	}
	return s
}

// Document returns a goquery Document instance of the document content.
func Document(ctx playwright.BrowserContext, url string) (*model.Document, error) {
	page, err := NewPage(ctx)
	if err != nil {
		return nil, err
	}
	defer func(page playwright.Page) {
		_ = page.Close()
	}(page)

	var resp playwright.Response
	if resp, err = GoTo(page, url); err != nil {
		return nil, err
	} else if IsRequestBlocked(resp) || IsPageBlocked(page) {
		return nil, errors.New("blocked")
	}

	var s string
	if s, err = page.Content(); err != nil {
		return nil, err
	}

	return model.ParseDocument(s)
}

// Meta returns a map of meta tags by inspecting meta tag properties.
func Meta(page playwright.Page) map[string]string {

	m := make(map[string]string)

	var k, v string
	for _, l := range Locators(page, "meta") {
		if k = attribute(l, "name"); k == "" {
			if k = attribute(l, "property"); k == "" {
				continue
			}
		}
		if v = attribute(l, "content"); v == "" {
			continue
		}
		m[k] = v
	}

	return m
}

// Links returns absolute and relative links of a document by inspecting anchor tag properties.
// - ✅ (absolute) https://ByteLyon.com/dashboard
// - ✅ (relative) /dashboard
// - ❌ (fragment) #contact
// - ❌ (download) foo.pdf
// - ❌ (schemes) mailto:foo@bar.com
// - ❌ (js) javascript:void(0);
func Links(page playwright.Page) []string {
	var m = make(map[string]bool)

	for _, a := range Locators(page, "a") {

		// does it have a hypertext reference?
		href, err := a.GetAttribute("href")
		if err != nil {
			continue
		}

		// trim whitespace (yes, technically it's possible)
		if href = strings.TrimSpace(href); href == "" {
			continue
		}

		// is it a js link?
		if strings.Contains(href, "javascript:") {
			continue
		}

		// is it a file link?
		if HasFileExtension(href) {
			continue
		}

		// is it a fragment?
		if strings.HasPrefix(href, "#") {
			continue
		}

		// is it a browser function?
		if hrefSchemes.MatchString(href) {
			continue
		}

		m[href] = true
	}

	return slices.Collect(maps.Keys(m))
}

// Paragraphs returns a list of unique paragraphs in the document.
func Paragraphs(page playwright.Page) (paragraphs []string) {

	uniqueParagraphs := make(map[string]int)
	for i, p := range Locators(page, "p") {
		//if txt := textContent(p); txt != "" && !parser.Skip(txt) {
		//	uniqueParagraphs[txt] = i
		//}
		if txt := textContent(p); txt != "" {
			uniqueParagraphs[txt] = i
		}
	}

	orderedParagraphs := make(map[int]string)
	for k, v := range uniqueParagraphs {
		orderedParagraphs[v] = k
	}

	for _, k := range slices.Sorted(maps.Keys(orderedParagraphs)) {
		paragraphs = append(paragraphs, orderedParagraphs[k])
	}

	return
}

// Headings returns a map of headings by their level.
func Headings(page playwright.Page) (headings map[string][]string) {
	for _, h := range []string{"h1", "h2", "h3", "h4", "h5", "h6"} {
		for _, l := range Locators(page, h) {
			headings[h] = append(headings[h], textContent(l))
		}
	}
	return
}

func SearchGoogle(q string, ctx playwright.BrowserContext) (page playwright.Page, err error) {
	if page, err = NewPage(ctx); err != nil {
		return
	}

	var resp playwright.Response
	if resp, err = GoTo(page, "https://www.google.com"); err != nil {
		return
	}

	if IsRequestBlocked(resp) || IsPageBlocked(page) {
		if WaitForLoadState(page); IsRequestBlocked(resp) || IsPageBlocked(page) {
			return
		}
	}

	if err = Click(page, googleSearchInputSelectors...); err != nil {
		return
	} else if err = Type(page, q); err != nil {
		return
	} else if err = Press(page, "Enter"); err != nil {
		return
	} else if err = WaitForLoadState(page); err != nil {
		return
	} else if IsPageBlocked(page) {
		return
	}

	log.Info().Msgf("Reached Google SERP for query: %s", q)

	return
}

func Locators(page playwright.Page, s string) []playwright.Locator {
	arr, err := page.Locator(s).All()
	if err != nil {
		log.Warn().
			Err(err).
			Str("selector", s).
			Msg("failed to get locators")
		return []playwright.Locator{}
	}
	return arr
}

func textContent(l playwright.Locator) string {
	s, err := l.TextContent()
	if err != nil {
		log.Warn().Err(err).Msg("failed to get text content")
		return ""
	}
	return strings.TrimSpace(s)
}

func attribute(l playwright.Locator, a string) string {
	s, err := l.GetAttribute(a)
	if err != nil {
		log.Warn().Err(err).Msg("failed to get attribute")
		return ""
	}
	return strings.TrimSpace(s)
}
