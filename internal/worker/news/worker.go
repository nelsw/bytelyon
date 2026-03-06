package news

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/nelsw/bytelyon/internal/client/fetch"
	"github.com/nelsw/bytelyon/pkg/db"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/rs/zerolog/log"
)

var (
	bingRegexp = regexp.MustCompile("</?News(:\\w+)>")
)

type Worker struct {
	*model.News

}

func New(bot *model.News) *Worker {
	return &Worker{bot}
}

func (w *Worker) Work() {

	q := strings.ReplaceAll(w.Target, ` `, `+`)
	urls := []string{
		fmt.Sprintf("https://news.google.com/rss/search?q=%s&hl=en-US&gl=US&ceid=US:en", q),
		fmt.Sprintf("https://www.bing.com/news/search?format=rss&q=%s", q),
		fmt.Sprintf("https://www.bing.com/search?format=rss&q=%s", q),
	}

	for _, url := range urls {
		w.workUrl(url)
	}
}

func (w *Worker) workUrl(url string) {

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

			log.Trace().Any("item", i).Msg("Processing RSS item")

			// if this job is brand new, save all the articles found
			// else persist articles published after the last update

			if w.UserID.Timestamp().Sub(w.UpdatedAt) !=  &&
				time.Time(*i.Time).Before(w.UpdatedAt) {
				log.Debug().Msgf("Skipping old article %s", i.Title)
				return
			}

			// check article data for blacklisted keywords
			titleParts := strings.Split(i.Title, " ")
			sourceParts := strings.Split(i.Source, " ")
			parts := append(titleParts, sourceParts...)
			for _, p := range parts {
				if _, ok := w.Rules[p]; ok {
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

			// scrub the source off the title and use it if the item source is blank
			if l, r, ok := strings.Cut(i.Title, " - "); ok {
				i.Title = l
				if i.Source == "" {
					i.Source = r
				}
			}

			// check if the description is HTML
			if idx := strings.Index(i.Description, `</a>`); idx > 0 {
				i.Description = i.Description[:idx]
				i.Description = i.Description[strings.LastIndex(i.Description, ">")+1:]
			}

			w.Items[model.URL(i.URL)] = model.Item{
				URL:         model.URL(i.URL),
				Title:       i.Title,
				Source:      i.Source,
				Description: i.Description,
				Published:   time.Time(*i.Time),
				CreatedAt:   time.Now(),
			}

			if err != nil {
				log.Warn().Err(err).Msg("failed to save news article")
			}
		})
	}
	wg.Wait()

	err = db.Put(w.News)
}
