package news

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/nelsw/bytelyon/internal/rss"
	"github.com/nelsw/bytelyon/pkg/entity"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

type Prowler struct {

	// ctx is the context of the browser, which is used to run the browser and the page
	ctx playwright.BrowserContext

	*entity.News

	lastProwl time.Time

	blackMap map[string]bool
}

func New(bot *model.Bot, ctx playwright.BrowserContext) *Prowler {
	return &Prowler{
		ctx:       ctx,
		News:      new(entity.News).From(bot.UserID, bot.Target),
		lastProwl: bot.WorkedAt,
		blackMap:  bot.BlackMap(),
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

	var exclude = make(map[string]bool)
	for _, a := range p.News.Articles {
		exclude[a.URL] = true
	}

	var items []*rss.Item
	for _, url := range urls {
		items = append(items, rss.Items(url, p.lastProwl, exclude)...)
	}

	var wg sync.WaitGroup
	for _, i := range items {
		wg.Go(func() {

			page := new(entity.Page).Scrape(i.Link, p.ctx)
			if page == nil {
				return
			}
			page.Save()

			a := page.MakeArticle(i.PublishedAt, i.Source, i.Description)
			for _, word := range a.Words() {
				if _, ok := p.blackMap[word]; ok {
					return
				}
			}

			p.Articles[a.URL] = a
			p.Save()
		})
	}
	wg.Wait()

	log.Info().Msgf("processed news worker %s", p.Topic)
}
