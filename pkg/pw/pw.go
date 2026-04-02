package pw

import (
	"github.com/nelsw/bytelyon/pkg/client"
	"github.com/nelsw/bytelyon/pkg/logs"
	"github.com/playwright-community/playwright-go"
)

var (
	Client *playwright.Playwright
)

func Init() {
	if err := playwright.Install(&playwright.RunOptions{Logger: logs.NewSlog()}); err != nil {
		panic(err)
	} else if Client, err = playwright.Run(); err != nil {
		panic(err)
	}
}

func NewBrowser(headless bool) (playwright.Browser, error) {
	return client.NewBrowser(Client, headless)
}

func NewBrowsers() (headless playwright.Browser, headed playwright.Browser, err error) {
	if headless, err = client.NewBrowser(Client, true); err != nil {
		return
	}
	if headed, err = client.NewBrowser(Client, false); err != nil {
		headless.Close()
	}
	return
}
