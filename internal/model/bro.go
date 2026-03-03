package model

import (
	. "github.com/nelsw/bytelyon/internal/util"
	. "github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

type Bro struct {
	*Playwright
	Browser
}

func NewBro(headless bool) (*Bro, error) {

	log.Trace().Msg("creating new Bro")
	var err error
	var bro = new(Bro)

	if bro.Playwright, err = Run(); err != nil {
		log.Err(err).Msg("failed to run Bro")
		return nil, err
	}

	bro.Browser, err = bro.Chromium.Launch(BrowserTypeLaunchOptions{
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
		log.Err(err).Msg("failed to create Bro")
		return nil, err
	}

	log.Debug().Msg("created Bro")

	return bro, nil
}
