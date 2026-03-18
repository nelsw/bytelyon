package sitemap

import (
	"time"

	"github.com/nelsw/bytelyon/pkg/db"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/rs/zerolog/log"
)

type Worker struct {
	*model.Bot
}

func New(bot *model.Bot) *Worker {
	return &Worker{bot}
}

func (w *Worker) Work() {

	m := NewMapper(&fetcher{}, w.Target)
	m.Add()
	m.Map(w.Target, 3)
	m.Wait()

	err := db.PutItem(w.NewBotResult(
		"relative", m.Relative(),
		"remote", m.Remote(),
	))

	log.Err(err).Msg("put sitemap result")

	w.WorkedAt = time.Now().UTC()
	if err = db.PutItem(w); err != nil {
		log.Err(err).Msg("Failed to update sitemap")
	}
}
