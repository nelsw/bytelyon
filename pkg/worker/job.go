package worker

import (
	"time"

	"github.com/nelsw/bytelyon/pkg/db"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/news"
	"github.com/nelsw/bytelyon/pkg/search"
	"github.com/nelsw/bytelyon/pkg/sitemap"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

type Job struct {
	bro playwright.Browser
	ctx playwright.BrowserContext
	bot *model.Bot
}

func NewJob(bro playwright.Browser, ctx playwright.BrowserContext, bot *model.Bot) *Job {
	return &Job{
		bro: bro,
		ctx: ctx,
		bot: bot,
	}
}

func (j *Job) Close() {
	if err := j.ctx.Close(); err != nil {
		log.Err(err).Msg("failed to close browser context")
	}
	if err := j.bro.Close(); err != nil {
		log.Err(err).Msg("failed to close browser")
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
		log.Warn().Err(err).Msg("Failed to Save Search Bot")
	}
}

func (j *Job) Work() {

	defer j.Close()

	switch j.bot.Type {
	case model.SearchBotType:
		search.Work(j.ctx, j.bot.UserID, j.bot.Target, j.bot.BlackMap())
	case model.SitemapBotType:
		sitemap.Work(j.ctx, j.bot.UserID, j.bot.Target)
	case model.NewsBotType:
		news.Work(j.ctx, j.bot.UserID, j.bot.Target, j.bot.BlackMap(), j.bot.WorkedAt)
	default:
		log.Warn().Msgf("bot type [%s] not supported", j.bot.Type)
	}
}
