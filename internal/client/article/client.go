package article

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/nelsw/bytelyon/internal/model"
	"github.com/rs/zerolog/log"
)

type Client struct {
	url   string
	after time.Time
}

func NewClient(query string, after time.Time) *Client {
	return &Client{
		fmt.Sprintf("https://news.google.com/rss/search?q=%s", strings.ReplaceAll(query, ` `, `+`)),
		after,
	}
}

func (c *Client) Fetch() ([]*model.Article, error) {

	rss, err := NewRSS(c.url)
	if err != nil {
		return nil, err
	}

	var articles []*model.Article

	var wg sync.WaitGroup
	for _, i := range rss.Channel.Items {

		wg.Go(func() {

			t := time.Time(*i.Time)
			if t.Before(c.after) {
				return
			}

			var u string
			if u, err = decodeURL(i.URL); err != nil {
				log.Warn().Err(err).Send()
				u = i.URL
			}

			articles = append(articles, &model.Article{
				URL:       u,
				Title:     strings.TrimSuffix(i.Title, " - "+i.Source),
				Source:    i.Source,
				Published: t,
			})
		})
	}
	wg.Wait()

	return articles, nil
}
