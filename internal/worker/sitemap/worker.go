package sitemap

import (
	"sort"

	"github.com/nelsw/bytelyon/internal/db"
	"github.com/nelsw/bytelyon/internal/model"
	"github.com/nelsw/bytelyon/internal/util"
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

	db.Create(&model.Sitemap{
		Bot:      w.Bot,
		URL:      w.Target,
		Domain:   util.Domain(w.Target),
		Relative: m.Relative(),
		Remote:   m.Remote(),
	})
}
