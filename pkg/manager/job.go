package manager

import (
	"time"

	"github.com/nelsw/bytelyon/pkg/db"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

type Job struct {
	ctx playwright.BrowserContext
	bot *model.Bot
}

func NewJob(bot *model.Bot, ctx ...playwright.BrowserContext) *Job {
	var j = new(Job)
	j.bot = bot
	if len(ctx) > 0 {
		j.ctx = ctx[0]
	}
	return j
}

func (j *Job) Work() {
	switch j.bot.Type {
	case "search":
		j.doSearch()
	case "sitemap":
		j.doSitemap()
	case "news":
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

	// save bot
	if err := db.PutItem(j.bot); err != nil {
		log.Warn().Err(err).Msg("Failed to Save Search Bot (DB)")
	}
}
