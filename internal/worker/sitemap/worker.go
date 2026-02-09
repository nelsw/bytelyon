package sitemap

import (
	"sort"

	"github.com/nelsw/bytelyon/internal/model"
	"github.com/nelsw/bytelyon/internal/util"
	"gorm.io/gorm"
)

type Worker struct {
	*gorm.DB
	*model.Bot
}

func New(db *gorm.DB, b *model.Bot) *Worker {
	return &Worker{db, b}
}

func (w *Worker) Work() {

	m := NewMapper(&fetcher{}, w.Target)
	m.Add()
	m.Map(w.Target, 3)
	m.Wait()

	sort.Strings(m.Relative())
	sort.Strings(m.Remote())

	w.Create(&model.Sitemap{
		Bot:      w.Bot,
		URL:      w.Target,
		Domain:   util.Domain(w.Target),
		Relative: m.Relative(),
		Remote:   m.Remote(),
	})
}
