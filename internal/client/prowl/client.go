package prowl

import (
	"errors"
	"log/slog"
	"regexp"

	"github.com/nelsw/bytelyon/internal/config"
	"github.com/nelsw/bytelyon/internal/util"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
	slogzerolog "github.com/samber/slog-zerolog/v2"
)

var (
	blockedRegex = regexp.MustCompile("(google.com/sorry|captcha|unusual traffic)")
)

type Client struct {
	Headless *bool
	*playwright.Playwright
	playwright.Browser
	playwright.BrowserContext
}

func Init() {
	var sl slog.Level
	if config.IsReleaseMode() {
		sl = slog.LevelError
	} else if config.IsDebugMode() {
		sl = slog.LevelInfo
	} else {
		sl = slog.LevelDebug
	}
	err := playwright.Install(&playwright.RunOptions{
		Logger: slog.New(slogzerolog.Option{
			Level:  sl,
			Logger: util.Ptr(log.Logger),
		}.NewZerologHandler()),
	})
	if err != nil {
		panic(err)
	}
}

func New(headless bool) (c *Client, err error) {

	c = &Client{Headless: &headless}

	if c.Playwright, err = playwright.Run(); err != nil {
		return
	} else if err = c.NewBrowser(); err != nil {
		return
	} else if err = c.NewBrowserContext(); err != nil {
		return
	}
	return
}

func (c *Client) IsBlocked(aa ...any) error {
	for _, a := range aa {
		switch t := a.(type) {
		case playwright.Page:
			if blockedRegex.MatchString(t.URL()) {
				return errors.New("blocked: " + t.URL())
			}
		case playwright.Response:
			if t.Status() >= 400 {
				return errors.New("blocked: " + t.URL())
			}
		}
	}
	return nil
}
