package sitemap

import (
	"sort"

	"github.com/nelsw/bytelyon/internal/model"
	"github.com/nelsw/bytelyon/internal/service/db"
	"github.com/nelsw/bytelyon/internal/util"
	"github.com/rs/zerolog/log"
)

type Worker struct {
	*model.SitemapBot
}

func New(bot *model.SitemapBot) *Worker {
	return &Worker{bot}
}

func (w *Worker) Work() {

	m := NewMapper(&fetcher{}, w.Target)
	m.Add()
	m.Map(w.Target, 3)
	m.Wait()

	sort.Strings(m.Relative())
	sort.Strings(m.Remote())

	err := db.Save(&model.SitemapBotData{
		BotID:    w.BotID,
		URL:      w.Target,
		Domain:   util.Domain(w.Target),
		Relative: m.Relative(),
		Remote:   m.Remote(),
	})

	if err != nil {
		log.Err(err).Msg("Failed to create sitemap")
	}
}
