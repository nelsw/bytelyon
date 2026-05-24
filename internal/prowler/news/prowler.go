package news

import (
	"fmt"
	"strings"

	"github.com/nelsw/bytelyon/internal/rss"
	"github.com/nelsw/bytelyon/pkg/entity"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

type Prowler struct {
	*entity.News

	// ctx is the context of the browser, which is used to run the browser and the page
	ctx playwright.BrowserContext
}

func New(e *entity.News, ctx playwright.BrowserContext) *Prowler {
	return &Prowler{
		ctx:  ctx,
		News: e,
	}
}

func (p *Prowler) Prowl() {

	log.Info().Msgf("processing news worker %s", p.Topic)

	q := strings.ReplaceAll(p.Topic, " ", "+")
	urls := []string{
		fmt.Sprintf("https://www.bing.com/search?format=rss&q=%s", q),
		fmt.Sprintf("https://www.bing.com/news/search?format=rss&q=%s", q),
		fmt.Sprintf("https://news.google.com/rss/search?q=%s&hl=en-US&gl=US&ceid=US:en", q),
	}

	var items []*rss.Item
	for _, url := range urls {
		items = append(items, rss.Items(url, p.After, p.Exclude)...)
	}

	log.Info().Msgf("processed news worker %s", p.Topic)
}
