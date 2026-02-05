package prowl

import (
	. "github.com/nelsw/bytelyon/internal/util"

	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

func (c *Client) NewBrowser() (err error) {

	c.Browser, err = c.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: c.Headless,
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

	log.Err(err).Msg("Client - NewBrowser")

	return
}

func (c *Client) Close() {
	if err := c.BrowserContext.Close(); err != nil {
		log.Warn().Err(err).Msg("Failed to close Client Context")
	} else if err = c.Browser.Close(); err != nil {
		log.Warn().Err(err).Msg("Failed to close Client Browser")
	}
}
