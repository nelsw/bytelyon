package sitemap

import (
	"context"
	"sort"

	"github.com/nelsw/bytelyon/internal/db"
	"github.com/nelsw/bytelyon/internal/model"
	"github.com/nelsw/bytelyon/internal/util"
	"github.com/rs/zerolog/log"
)

type Worker struct {
	*model.Bot
}

func New(b *model.Bot) *Worker {
	return &Worker{b}
}

func (w *Worker) Work() {

	m := NewMapper(&fetcher{}, w.Target)
	m.Add()
	m.Map(w.Target, 3)
	m.Wait()

	sort.Strings(m.Relative())
	sort.Strings(m.Remote())

	err := db.Builder[model.Sitemap]().Create(context.Background(), &model.Sitemap{
		BotID:    w.Bot.ID,
		URL:      w.Target,
		Domain:   util.Domain(w.Target),
		Relative: m.Relative(),
		Remote:   m.Remote(),
	})

	if err != nil {
		log.Err(err).Msg("Failed to create sitemap")
	}
}
