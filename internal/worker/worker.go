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
		w.doArticleWork()
	case model.SitemapType:
		w.doSitemapWork()
	case model.SearchType:
		w.doSearchWork()
	default:
		log.Warn().Msg("unknown job type")
	}
}

func (w worker) doArticleWork() {
	if arr := article.New(w.Job).Work(); len(arr) > 0 {
		w.Save(arr)
	}
}

func (w worker) doSitemapWork() {
	if m := sitemap.New(w.Job).Work(); m != nil {
		w.Save(m)
	}
}

func (w worker) doSearchWork() {
	if a := search.New(w.Job).Work(); a != nil {
		w.Save(a)
	}
}
