package worker

import (
	"github.com/nelsw/bytelyon/internal/client/article"
	"github.com/nelsw/bytelyon/internal/model"
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
	arr, err := article.NewClient(w.Target, w.UpdatedAt).Fetch()
	if err != nil {
		log.Error().Err(err).Msg("failed to fetch article")
		return
	}
	w.db.Save(arr)
}

func (w worker) doSitemapWork() {

}

func (w worker) doSearchWork() {

}
