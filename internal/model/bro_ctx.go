package model

import (
	"math/rand"

	. "github.com/nelsw/bytelyon/internal/util"
	. "github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

var (
	agents = []string{
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
	broCtxInitScript = `() => {
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
}`
)

type BroCtx struct {
	BrowserContext
}

func NewBroCtx(bro *Bro, bcs BroCtxState) (*BroCtx, error) {

	log.Trace().Msg("creating new playwright context")

	var err error
	var ctx = new(BroCtx)

	ctx.BrowserContext, err = bro.NewContext(BrowserNewContextOptions{
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
		UserAgent:         Ptr(agents[rand.Intn(len(agents))]),
		StorageState:      bcs.StorageState(),
	})

	if err != nil {
		log.Err(err).Msg("failed to create new playwright context")
		return nil, err
	}

	err = ctx.AddInitScript(Script{Content: Ptr(broCtxInitScript)})

	if err != nil {
		log.Err(err).Msg("failed to add init script to playwright context")
		return nil, err
	}

	ctx.SetDefaultTimeout(60_000)

	log.Debug().Msg("created new playwright context")

	return ctx, nil
}
