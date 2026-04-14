package pw

import (
	"github.com/nelsw/bytelyon/pkg/client"
	"github.com/nelsw/bytelyon/pkg/logs"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

var (
	Client *playwright.Playwright
)

func Init() {
	log.Debug().Msg("installing playwright drivers")
	if err := playwright.Install(&playwright.RunOptions{Logger: logs.NewSlog()}); err != nil {
		log.Panic().Err(err).Msg("failed to install playwright")
	}

	log.Debug().Msg("starting a playwright instance")
	var err error
	if Client, err = playwright.Run(); err != nil {
		log.Panic().Err(err).Msg("failed to start playwright")
	}

	log.Info().Msg("playwright initialized successfully")
}

func NewBrowser(headless bool) (playwright.Browser, error) {
	return client.NewBrowser(Client, headless)
}
