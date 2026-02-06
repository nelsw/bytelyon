package sitemap

import (
	"github.com/nelsw/bytelyon/internal/model"
	"github.com/nelsw/bytelyon/internal/util"
)

type Worker struct {
	*model.Job
}

func New(j *model.Job) *Worker {
	return &Worker{j}
}

func (w *Worker) Work() *model.Sitemap {

	m := NewMapper(&fetcher{}, w.Target)
	m.Add()
	m.Map(w.Target, 3)
	m.Wait()

	return &model.Sitemap{
		JobID:    w.ID,
		URL:      w.Target,
		Domain:   util.Domain(w.Target),
		Relative: m.Relative(),
		Remote:   m.Remote(),
	}
}
