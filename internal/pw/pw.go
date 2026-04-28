package pw

import (
	"math/rand"
	"regexp"

	"github.com/nelsw/bytelyon/pkg/logs"
	. "github.com/nelsw/bytelyon/pkg/util"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

var (
	blockedRegex = regexp.MustCompile("(google.com/sorry|captcha|unusual traffic)")
	installed    = false
)

// Install gets drivers from the web and installs them to the machine running ByteLyon
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
			"--disable-web-security",
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

// Press executes a keyboard event on a page.
func Press(page playwright.Page, s string) error {
	return page.Keyboard().Press(s, playwright.KeyboardPressOptions{
		Delay: Ptr(Between(200, 500.0)),
	})
}

// NewPage creates a new page in the browser context.
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

// Click the first page element located by the given selectors.
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

// Content returns the page content or an empty string if the page has failed to load.
func Content(page playwright.Page) string {
	s, err := page.Content()
	if err != nil {
		log.Err(err).Msg("failed to get page content")
		return ""
	}
	return s
}

// Screenshot returns the screenshot of the page as a byte array or an empty byte array if the page has failed to load.
func Screenshot(page playwright.Page, opts ...playwright.PageScreenshotOptions) []byte {
	opts = append(opts, playwright.PageScreenshotOptions{FullPage: Ptr(true)})
	b, err := page.Screenshot()
	if err != nil {
		return nil
	}
	return b
}

// Title returns the page title or an empty string if the page has failed to load.
func Title(page playwright.Page) string {
	s, err := page.Title()
	if err != nil {
		return ""
	}
	return s
}
