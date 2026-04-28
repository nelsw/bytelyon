package worker

import (
	"github.com/nelsw/bytelyon/pkg/service/sitemaps"
	"github.com/rs/zerolog/log"
)

func (j *Job) doSitemap() {
	if err := sitemaps.Create(j.bot.Target, 5, j.ctx); err != nil {
		log.Err(err).Msg("failed to create sitemap")
	}
}
