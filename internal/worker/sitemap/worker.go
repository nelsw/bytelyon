package sitemap

import (
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/nelsw/bytelyon/internal/model"
	"github.com/nelsw/bytelyon/internal/service/db"
	"github.com/rs/zerolog/log"
)

type Worker struct {
	*model.BotSitemap
}

func New(bot *model.BotSitemap) *Worker {
	return &Worker{bot}
}

func (w *Worker) Work() {

	m := NewMapper(&fetcher{}, w.Target)
	m.Add()
	m.Map(w.Target, 3)
	m.Wait()

	sort.Strings(m.Relative())
	sort.Strings(m.Remote())

	err := db.Save(&model.BotSitemapResult{
		Model:    model.Make(w.UserID),
		ID:       uuid.Must(uuid.NewV7()),
		Target:   w.Target,
		Relative: m.Relative(),
		Remote:   m.Remote(),
	})

	if err != nil {
		log.Err(err).Msg("Failed to create sitemap")
	}

	w.Bot.UpdatedAt = time.Now()

	err = db.Save(w.BotSitemap)
}
