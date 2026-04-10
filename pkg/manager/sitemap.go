package manager

import (
	"github.com/nelsw/bytelyon/pkg/service/pages"
	"github.com/nelsw/bytelyon/pkg/service/sitemaps"
	"github.com/rs/zerolog/log"
)

func (j *Job) doSitemap() {

	m, err := sitemaps.Create(j.bot.Target, 3)

	if err != nil {
		log.Err(err).Msg("failed to create sitemap")
		return
	}

	for _, url := range m.URLs.Slice() {
		if err = pages.Create(url, j.ctx); err != nil {
			log.Err(err).Msg("failed to create sitemap page")
		}
	}
}
