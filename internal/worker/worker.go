package worker

import (
	"github.com/nelsw/bytelyon/internal/model"
	"github.com/nelsw/bytelyon/internal/worker/article"
	"github.com/nelsw/bytelyon/internal/worker/search"
	"github.com/nelsw/bytelyon/internal/worker/sitemap"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type Worker interface {
	Work()
}

type worker struct {
	*gorm.DB
	*model.Job
}

func New(db *gorm.DB, job *model.Job) Worker {
	return &worker{db, job}
}

func (w worker) Work() {
	switch w.Type {
	case model.ArticleType:
		if arr := article.New(w.Job).Work(); len(arr) > 0 {
			w.Save(arr)
		}
	case model.SitemapType:
		if m := sitemap.New(w.Job).Work(); m != nil {
			w.Save(m)
		}
	case model.SearchType:
		if a := search.New(w.Job).Work(); a != nil {
			w.Save(a)
		}
	default:
		log.Warn().Msg("unknown job type")
		return
	}
	// a frequency of 1 means to only work the job once;
	// it's been worked, reset the frequency to 0 (pause).
	if w.Frequency == 1 {
		w.Frequency = 0
	}
	w.Save(w.Job)
}
