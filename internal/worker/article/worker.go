package article

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/nelsw/bytelyon/internal/client/fetch"
	"github.com/nelsw/bytelyon/internal/model"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type Worker struct {
	*gorm.DB
	*model.Bot
}

func New(db *gorm.DB, b *model.Bot) *Worker {
	return &Worker{db, b}
}

func (c *Worker) Work() {

	query := strings.ReplaceAll(c.Target, ` `, `+`)
	url := fmt.Sprintf("https://news.google.com/rss/search?q=%s", query)

	var rss RSS
	if err := fetch.New(url).XML(&rss); err != nil {
		log.Err(err).Send()
		return
	}

	var wg sync.WaitGroup
	for _, i := range rss.Channel.Items {

		wg.Go(func() {

			if time.Time(*i.Time).Before(time.Unix(int64(c.UpdatedAt), 0)) {
				return
			}

			u, err := decodeURL(i.URL)
			if err != nil {
				log.Warn().Err(err).Send()
				u = i.URL
			}

			title := strings.TrimSuffix(i.Title, " - "+i.Source)
			parts := strings.Split(title, " ")
			for _, p := range parts {
				if _, ok := c.Ignore()[p]; ok {
					continue
				}
			}

			c.Create(&model.Article{
				BotID:     c.Bot.ID,
				URL:       u,
				Title:     title,
				Source:    i.Source,
				Published: time.Time(*i.Time),
			})
		})
	}
	wg.Wait()
}
