package news

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/nelsw/bytelyon/internal/client/fetch"
	"github.com/nelsw/bytelyon/internal/db"
	"github.com/nelsw/bytelyon/internal/model"
	"github.com/rs/zerolog/log"
)

var (
	bingRegexp = regexp.MustCompile("</?News(:\\w+)>")
)

type Worker struct {
	*model.Bot
}

func New(b *model.Bot) *Worker {
	return &Worker{b}
}

func (c *Worker) Work() {

	q := strings.ReplaceAll(c.Target, ` `, `+`)
	urls := []string{
		fmt.Sprintf("https://news.google.com/rss/search?q=%s&hl=en-US&gl=US&ceid=US:en", q),
		fmt.Sprintf("https://www.bing.com/news/search?format=rss&q=%s", q),
		fmt.Sprintf("https://www.bing.com/search?format=rss&q=%s", q),
	}

	for _, url := range urls {
		c.workUrl(url)
	}

}

func (c *Worker) workUrl(url string) {

	b, err := fetch.New(url).Bytes()
	if err != nil {
		log.Err(err).Str("url", url).Msg("Failed to fetch RSS feed")
		return
	}

	if strings.Contains(url, "bing.com") {
		b = []byte(bingRegexp.ReplaceAllStringFunc(string(b), func(s string) string {
			return strings.ReplaceAll(s, ":", "_")
		}))
	}

	var rss RSS
	if err = xml.Unmarshal(b, &rss); err != nil {
		log.Err(err).Str("url", url).Msg("Failed to unmarshal RSS feed")
		return
	}

	var wg sync.WaitGroup
	for _, i := range rss.Channel.Items {

		wg.Go(func() {

			// if this job is brand new, save all the articles found
			// else persist articles published after the last update
			if c.CreatedAt != c.UpdatedAt &&
				time.Time(*i.Time).Before(time.Unix(int64(c.UpdatedAt), 0)) {
				log.Debug().Msgf("Skipping old article %s", i.Title)
				return
			}

			// check article data for blacklisted keywords
			titleParts := strings.Split(i.Title, " ")
			sourceParts := strings.Split(i.Source, " ")
			parts := append(titleParts, sourceParts...)
			for _, p := range parts {
				if _, ok := c.Ignore()[p]; ok {
					log.Info().Msgf("Skipping blacklisted article %s", p)
					return
				}
			}

			// work some magic to circumvent Googles bot protection
			if strings.Contains(url, "google.com") {
				if u, decodeErr := decodeURL(i.URL); decodeErr != nil {
					log.Warn().Err(decodeErr).Send()
				} else {
					i.URL = u
				}
			}

			// scrub the source off the title and
			// use it if the item source is blank
			if title, source, ok := strings.Cut(i.Title, " - "); ok && i.Source == "" {
				i.Source = source
				i.Title = title
			}

			err = db.Create(&model.News{
				Bot:         c.Bot,
				URL:         i.URL,
				Title:       i.Title,
				Source:      i.Source,
				Published:   time.Time(*i.Time),
				Description: i.Description,
			})

			if err != nil {
				log.Warn().Err(err).Msg("failed to save news article")
			}
		})
	}
	wg.Wait()
}
