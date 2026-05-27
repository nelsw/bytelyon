package news

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/nelsw/bytelyon/internal/pw"
	"github.com/nelsw/bytelyon/pkg/document"
	"github.com/nelsw/bytelyon/pkg/https"
	"github.com/nelsw/bytelyon/pkg/id"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/page"
	"github.com/nelsw/bytelyon/pkg/urls"
	"github.com/oklog/ulid/v2"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

func Work(ctx playwright.BrowserContext, userID ulid.ULID, topic string, exclude map[string]bool, after time.Time) {

	headlines := fetch(topic, exclude, after)
	if len(headlines) == 0 {
		return
	}

	m := FindOrNew(userID, topic)

	var wg sync.WaitGroup
	for _, h := range headlines {
		wg.Go(routine(ctx, m, h))
	}
	wg.Wait()

	Save(userID, topic, m.Map)
}

func routine(ctx playwright.BrowserContext, m *model.SyncMap[string, *Headline], h *Headline) func() {
	return func() {
		if m.Has(h.URL) {
			return
		}

		content, screenshot := pw.Scrape(h.URL, ctx)
		if content == "" {
			return
		}
		doc := document.New(content)

		a := &Article{
			Body:        doc.Paragraphs,
			Description: doc.Meta.Description(),
			ID:          h.ID,
			Image:       doc.Meta.Image(),
			Keywords:    doc.Meta.Keywords(),
			Source:      doc.Meta.Source(),
			Title:       h.Title,
			URL:         h.URL,
		}

		if err := page.SaveObject(a.URL, a.ID, a); err != nil {
			log.Warn().Err(err).Msg("failed to save article object")
			return
		}

		if err := page.SaveScreenshot(a.URL, a.ID, screenshot); err != nil {
			log.Warn().Err(err).Msg("failed to save article screenshot")
			return
		}
		h.URL = ""
		m.Set(h.URL, h)
	}
}

func fetch(topic string, exclude map[string]bool, after time.Time) (headlines []*Headline) {
	q := strings.ReplaceAll(topic, " ", "+")

	arr := []string{
		fmt.Sprintf("https://news.google.com/rss/search?q=%s&hl=en-US&gl=US&ceid=US:en", q),
		fmt.Sprintf("https://www.bing.com/news/search?format=rss&q=%s", q),
		fmt.Sprintf("https://www.bing.com/search?format=rss&q=%s", q),
	}

	for _, u := range arr {
		b, err := https.Get(u)
		if err != nil {
			log.Err(err).Send()
			return
		}

		items := chomps(string(b), "<item>", "</item>")
		log.Debug().Str("url", u).Int("count", len(items)).Send()
		var pubDate, title, link string
		for _, item := range items {

			pubDate, item = chomp(item, "<pubDate>", "</pubDate>")
			ts, _ := time.Parse(time.RFC1123, pubDate)
			if ts.IsZero() {
				ts = time.Now()
			} else if ts.Before(after) {
				continue
			}

			title, item = chomp(item, "<title>", "</title>")
			for k := range exclude {
				if strings.Contains(title, k) {
					continue
				}
			}

			link, item = chomp(item, "<link>", "</link>")
			if strings.HasPrefix(link, "http://www.bing.com/news") {
				link = decodeBingNewsLink(link)
			} else if strings.HasPrefix(link, "https://news.google.com/rss/articles/") {
				link = decodeGoogleURL(link)
			}

			headlines = append(headlines, &Headline{
				id.New(ts),
				title,
				urls.Clean(link),
			})
		}
	}

	log.Debug().Int("count", len(headlines)).Msg("news")
	return
}
