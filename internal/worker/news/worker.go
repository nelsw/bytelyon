package news

import (
	"context"
	"encoding/xml"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	client "github.com/nelsw/bytelyon/internal/client/dynamodb"
	"github.com/nelsw/bytelyon/internal/client/fetch"
	"github.com/nelsw/bytelyon/internal/model"
	"github.com/rs/zerolog/log"
)

var (
	bingRegexp = regexp.MustCompile("</?News(:\\w+)>")
)

type Worker struct {
	context.Context
	*dynamodb.Client
	*model.NewsBot
}

func New(ctx context.Context, dbc *dynamodb.Client, bot *model.NewsBot) *Worker {
	return &Worker{ctx, dbc, bot}
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

			// if this job is brand new, save all the articles found
			// else persist articles published after the last update
			if time.Time(*i.Time).Before(w.UpdatedAt) {
				log.Debug().Msgf("Skipping old article %s", i.Title)
				return
			}

			// check article data for blacklisted keywords
			titleParts := strings.Split(i.Title, " ")
			sourceParts := strings.Split(i.Source, " ")
			parts := append(titleParts, sourceParts...)
			for _, p := range parts {
				if _, ok := w.Ignore()[p]; ok {
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

			err = client.PutItem(w.Context, w.Client, &model.NewsBotData{
				BotID:       w.BotID,
				URL:         i.URL,
				Title:       i.Title,
				Source:      i.Source,
				Description: i.Description,
				Published:   time.Time(*i.Time),
			})

			if err != nil {
				log.Warn().Err(err).Msg("failed to save news article")
			}
		})
	}
	wg.Wait()
}
