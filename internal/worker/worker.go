package worker

import (
	"github.com/nelsw/bytelyon/internal/model"
	"github.com/nelsw/bytelyon/internal/worker/article"
	"github.com/nelsw/bytelyon/internal/worker/sitemap"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type Worker interface {
	Work()
}

type worker struct {
	db *gorm.DB
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
	}
	log.Warn().Msg("unknown job type")
}

func (w worker) doArticleWork() {
	arr, err := article.New(w.Target, w.UpdatedAt).Work()
	if err != nil {
		log.Error().Err(err).Msg("failed to do article work")
		return
	}
	w.db.Save(arr)
}

func (w worker) doSitemapWork() {
	m := sitemap.New(w.Target).Work()
	if m == nil {
		log.Error().Msg("failed to do sitemap work")
		return
	}
	w.db.Save(m)
}

func (w worker) doSearchWork() {

}
