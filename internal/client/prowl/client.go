package prowl

import (
	"errors"
	"regexp"

	"github.com/nelsw/bytelyon/internal/logger"
	"github.com/nelsw/bytelyon/internal/util"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

var (
	blockedRegex = regexp.MustCompile("(google.com/sorry|captcha|unusual traffic)")
)

type Client struct {
	*playwright.Playwright
	playwright.Browser
	playwright.BrowserContext
}

func Init() {
	util.Check(playwright.Install(&playwright.RunOptions{Logger: logger.NewSlog()}))
}

func New(headless bool) (*Client, error) {

	var c = new(Client)
	var err error

	if c.Playwright, err = playwright.Run(); err != nil {
		return nil, err
	}
	log.Info().Msg("Client Initialized Playwright")

	if err = c.NewBrowser(headless); err != nil {
		return nil, err
	} else if err = c.NewBrowserContext(); err != nil {
		return nil, err
	}

	log.Info().Msg("Client Instantiated")

	return c, nil
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

func (c *Client) Close() {
	if err := c.BrowserContext.Close(); err != nil {
		log.Warn().Err(err).Msg("Failed to close Client Context")
	} else {
		log.Info().Msg("Client Context Closed")
	}

	if err := c.Browser.Close(); err != nil {
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
