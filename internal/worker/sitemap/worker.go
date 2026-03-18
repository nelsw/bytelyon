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

	err := db.PutItem(&model.BotResult{
		UserID: w.UserID,
		BotID:  w.ID,
		ID:     model.NewULID(),
		Type:   w.Type,
		Target: w.Target,
		Data: map[string]any{
			"relative": m.Relative(),
			"remote":   m.Remote(),
		},
	})

	if err != nil {
		log.Err(err).Msg("Failed to create sitemap")
	}

	w.WorkedAt = time.Now().UTC()
	if err = db.PutItem(w); err != nil {
		log.Err(err).Msg("Failed to update sitemap")
	}
}
