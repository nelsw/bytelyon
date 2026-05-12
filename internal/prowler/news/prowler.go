package news

import (
	"fmt"
	"strings"
	"sync"

	"github.com/nelsw/bytelyon/internal/pw"
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
}

func New(bot *model.Bot, ctx playwright.BrowserContext) *Prowler {
	return &Prowler{
		ctx:  ctx,
		News: new(entity.News).From(bot),
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
		if ii, err := rss.Items(url); err == nil {
			log.Debug().Str("url", url).Int("size", len(ii)).Msg("news items")
			items = append(items, ii...)
		}
	}

	ƒ := func(i *rss.Item) *entity.Page {

		l := log.With().
			Str("ƒ", "Prowler.put").
			Str("url", i.Link).
			Logger()

		l.Trace().Send()

		page, err := pw.NewPage(p.ctx)
		if err != nil {
			l.Warn().Msgf("NewPage failed: %s", err.Error())
			return nil
		}
		defer page.Close()

		if err = pw.Visit(page, i.Link); err != nil {
			l.Warn().Msgf("Visit failed: %s", err.Error())
			return nil
		}

		l.Debug().Send()

		return entity.NewPage(page)
	}

	var wg sync.WaitGroup
	for _, i := range items {
		wg.Go(func() {
			p.Add(ƒ(i), i.PublishedAt, i.Source, i.Description)
		})
	}
	wg.Wait()

	log.Info().Msgf("processed news worker %s", p.Topic)
}
