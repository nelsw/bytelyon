package article

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/nelsw/bytelyon/internal/client/fetch"
	"github.com/nelsw/bytelyon/internal/model"
	"github.com/rs/zerolog/log"
)

type Worker struct {
	url   string
	after time.Time
}

func New(query string, after time.Time) *Worker {
	return &Worker{
		fmt.Sprintf("https://news.google.com/rss/search?q=%s", strings.ReplaceAll(query, ` `, `+`)),
		after,
	}
}

func (c *Worker) Work() ([]*model.Article, error) {

	var rss RSS
	if err := fetch.New(c.url).XML(&rss); err != nil {
		return nil, err
	}

	var articles []*model.Article

	var wg sync.WaitGroup
	for _, i := range rss.Channel.Items {

		wg.Go(func() {

			if time.Time(*i.Time).Before(c.after) {
				return
			}

			u, err := decodeURL(i.URL)
			if err != nil {
				log.Warn().Err(err).Send()
				u = i.URL
			}

			articles = append(articles, &model.Article{
				URL:       u,
				Title:     strings.TrimSuffix(i.Title, " - "+i.Source),
				Source:    i.Source,
				Published: time.Time(*i.Time),
			})
		})
	}
	wg.Wait()

	return articles, nil
}
