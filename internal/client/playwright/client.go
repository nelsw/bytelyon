package playwright

import (
	"encoding/json"
	"math/rand"
	"os"
	"regexp"

	"github.com/nelsw/bytelyon/internal/logger"
	. "github.com/nelsw/bytelyon/internal/util"
	. "github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

var (
	blockedRegex = regexp.MustCompile("(google.com/sorry|captcha|unusual traffic)")
)

func Init() {
	Check(Install(&RunOptions{Logger: logger.NewSlog()}))
}

// New creates a new Playwright instance
func New() (c *Playwright, err error) {
	log.Trace().Msg("creating new playwright client")

	if c, err = Run(); err != nil {
		log.Err(err).Msg("failed to create playwright client")
		return nil, err
	}

	log.Debug().Msg("playwright client created")
	return c, nil
}

// NewBrowser creates a new Browser instance
func NewBrowser(c *Playwright, headless bool) (Browser, error) {

	log.Trace().Msg("creating playwright browser")

	bro, err := c.Chromium.Launch(BrowserTypeLaunchOptions{
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

	if err != nil {
		log.Err(err).Msg("failed to create playwright browser")
		return nil, err
	}

	log.Debug().Msg("created playwright browser")

	return bro, nil
}

// NewContext creates a new BrowserContext instance
func NewContext(bro Browser) (BrowserContext, error) {

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

	getState := func() *OptionalStorageState {

		var state OptionalStorageState

		b, err := os.ReadFile(BinDir("state.json"))
		if err != nil {
			log.Err(err).Msg("Client - Failed to read state.json")
			return &state
		}
		log.Debug().Msg("Client - Read state.json")

		if err = json.Unmarshal(b, &state); err != nil {
			log.Err(err).Msg("Client - Failed to unmarshal state.json")
			return &state
		}
		log.Debug().Msg("Client - Unmarshalled state.json")

		return &state
	}

	log.Trace().Msg("creating new playwright context")

	ctx, err := bro.NewContext(BrowserNewContextOptions{
		AcceptDownloads:   Ptr(true),
		ColorScheme:       ColorSchemeDark,
		ForcedColors:      ForcedColorsNone,
		HasTouch:          Ptr(false),
		IsMobile:          Ptr(false),
		JavaScriptEnabled: Ptr(true),
		Locale:            Ptr("en-US"),
		Permissions:       []string{"geolocation", "notifications"},
		ReducedMotion:     ReducedMotionNoPreference,
		TimezoneId:        Ptr("America/New_York"),
		UserAgent:         userAgent(),
		StorageState:      getState(),
	})

	if err != nil {
		log.Err(err).Msg("failed to create new playwright context")
		return nil, err
	}

	err = ctx.AddInitScript(Script{Content: Ptr(`() => {
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
		log.Err(err).Msg("failed to add init script to playwright context")
		return nil, err
	}

	ctx.SetDefaultTimeout(60_000)

	log.Debug().Msg("created new playwright context")

	return ctx, nil
}

// SetState writes the current BrowserContext to a file.
func SetState(ctx BrowserContext) {

	state, err := ctx.StorageState()
	if err != nil {
		log.Err(err).Msg("Client - Failed to get StorageState")
		return
	}
	log.Debug().Msg("Client - Got StorageState")

	var b []byte
	if b, err = json.Marshal(&state); err != nil {
		log.Err(err).Msg("Client - Failed to marshal StorageState")
		return
	}
	log.Debug().Msg("Client - Marshalled StorageState")

	if err = os.WriteFile(BinDir("state.json"), b, 0644); err != nil {
		log.Err(err).Msg("Client - Failed to write state.json")
		return
	}
	log.Debug().Msg("Client - Wrote state.json")
}

// Close closes the BrowserContext and Browser before stopping the Playwright instance.
func Close(c *Playwright, bro Browser, ctx BrowserContext) {
	if err := ctx.Close(); err != nil {
		log.Warn().Err(err).Msg("Failed to close Client Context")
	} else {
		log.Info().Msg("Client Context Closed")
	}

	if err := bro.Close(); err != nil {
		log.Warn().Err(err).Msg("Failed to close Client Browser")
	} else {
		log.Info().Msg("Client Browser Closed")
	}

	if err := c.Stop(); err != nil {
		log.Warn().Err(err).Msg("Failed to stop Client Playwright")
	} else {
		log.Info().Msg("Client Playwright Stopped")
	}
}

// IsPageBlocked determines if we have been blocked from visiting a URL.
func IsPageBlocked(page Page) bool {

	log.Trace().Msg("checking if page is blocked")

	if blockedRegex.MatchString(page.URL()) {
		log.Warn().Msg("page blocked")
		return true
	}

	log.Debug().Msg("page reached")
	return false
}

// IsRequestBlocked determines if we have been blocked from requesting a URL.
func IsRequestBlocked(res Response) bool {
	log.Trace().Msg("checking if request is blocked")

	if !res.Ok() {
		log.Warn().Msg("request blocked")
		return true
	}

	log.Debug().Msg("response ok")
	return false
}

// Type fills text to type into a focused element.
func Type(page Page, s string) error {
	log.Trace().Str("text", s).Msg("typing")
	err := page.Keyboard().Type(s, KeyboardTypeOptions{
		Delay: Ptr(Between(500.0, 1000.0)),
	})
	if err != nil {
		log.Err(err).Str("text", s).Msg("failed to type")
		return err
	}
	log.Debug().Str("text", s).Msg("typed")
	return err
}

// Press executes a keyboard event on a page.
func Press(page Page, s string) error {
	log.Trace().Str("key", s).Msg("pressing")
	err := page.Keyboard().Press(s, KeyboardPressOptions{
		Delay: Ptr(Between(200, 500.0)),
	})
	if err != nil {
		log.Err(err).Str("key", s).Msg("failed to press")
		return err
	}
	log.Debug().Str("key", s).Msg("pressed")
	return nil
}

// NewPage creates a new page in the browser context.
func NewPage(ctx BrowserContext) (Page, error) {

	log.Trace().Msg("creating new page")

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

	return page, nil
}

// GoTo returns the main resource response.
func GoTo(page Page, url string) (Response, error) {

	log.Trace().Str("url", url).Msg("go to url")

	res, err := page.Goto(url, PageGotoOptions{
		Timeout:   Ptr(10_000.0),
		WaitUntil: WaitUntilStateDomcontentloaded,
	})

	if err != nil {
		log.Err(err).Str("url", url).Msg("failed to go to url")
		return nil, err
	}

	log.Debug().Str("url", url).Msg("got to url")

	return res, nil
}

// Click the first page element located by the given selectors.
func Click(page Page, selectors ...string) error {

	log.Trace().Strs("selectors", selectors).Msg("clicking")

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

		log.Debug().Str("selector", selector).Msg("locator found")
		break
	}

	err := locator.Click(LocatorClickOptions{
		Delay: Ptr(Between(200, 500.0)),
	})

	if err != nil {
		log.Err(err).Str("selector", selector).Msg("failed to click locator")
		return err
	}

	log.Debug().Str("selector", selector).Msg("locator clicked")

	return nil
}

// WaitForLoadState returns nil when the required load state has been reached, or error if an exception occurred.
func WaitForLoadState(page Page, ls ...LoadState) error {

	s := LoadStateNetworkidle
	if len(ls) > 0 {
		s = &ls[0]
	}

	log.Trace().Any("state", s).Msg("waiting for load state")

	err := page.WaitForLoadState(PageWaitForLoadStateOptions{
		State:   s,
		Timeout: Ptr(60_000.0),
	})
	if err != nil {
		log.Err(err).Any("state", s).Msg("failed to reach load state")
		return err
	}

	log.Debug().Msg("load state reached")
	return nil
}
