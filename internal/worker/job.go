package worker

import (
	"time"

	"github.com/nelsw/bytelyon/internal/sitemap"
	"github.com/nelsw/bytelyon/pkg/db"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

type Job struct {
	ctx playwright.BrowserContext
	bot *model.Bot
}

func (j *Job) Work() {

	defer j.ctx.Close()

	switch j.bot.Type {
	case model.SearchBotType:
		j.doSearch()
	case model.SitemapBotType:
		sitemap.New(j.bot.Target, j.ctx).Work()
	case model.NewsBotType:
		j.doNews()
	default:
		log.Warn().Msgf("bot type [%s] not supported", j.bot.Type)
		return
	}

	// update bot worked at to now
	j.bot.WorkedAt = time.Now().UTC()

	// reset frequency if set to 1ns (once & stop)
	if j.bot.Frequency == 1 {
		j.bot.Frequency = 0
	}

	// update the storage state of the bot
	if state, err := j.ctx.StorageState(); err != nil {
		log.Warn().Err(err).Msg("Failed to get storage state")
	} else {
		j.bot.Fingerprint.SetState(state)
	}

	// save bot
	if err := db.Put(j.bot); err != nil {
		log.Warn().Err(err).Msg("Failed to Save Search Bot (DB)")
	}
}
