package worker

import (
	"time"

	"github.com/nelsw/bytelyon/internal/prowler/news"
	"github.com/nelsw/bytelyon/internal/prowler/search"
	"github.com/nelsw/bytelyon/internal/prowler/sitemap"
	"github.com/nelsw/bytelyon/pkg/db"
	"github.com/nelsw/bytelyon/pkg/em"
	"github.com/nelsw/bytelyon/pkg/entity"
	"github.com/nelsw/bytelyon/pkg/model"
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
		j.workSearch()
	case model.SitemapBotType:
		j.workSitemap()
	case model.NewsBotType:
		j.workNews()
	default:
		log.Warn().Msgf("bot type [%s] not supported", j.bot.Type)
	}
}

func (j *Job) workNews() {

	e := entity.NewNews(j.bot.UserID, j.bot.Target)

	if err := em.FindOrCreate(e); err != nil {
		log.Warn().Err(err).Msg("Failed to find or create sitemap")
		return
	}

	for _, s := range j.bot.BlackList {
		e.Exclude[s] = true
	}

	news.New(e, j.ctx).Prowl()
}

func (j *Job) workSearch() {

	e := entity.NewSearch(j.bot.UserID, j.bot.Target)

	if err := em.FindOrCreate(e); err != nil {
		log.Warn().Err(err).Msg("Failed to find or create sitemap")
		return
	}

	e.Exclude = j.bot.BlackMap()

	search.New(e, j.ctx).Prowl()
}

func (j *Job) workSitemap() {

	e := entity.NewSitemap(j.bot.UserID, j.bot.Target)

	if err := em.FindOrCreate(e); err != nil {
		log.Warn().Err(err).Msg("Failed to find or create sitemap")
		return
	}

	sitemap.New(e, j.ctx).Prowl()
}
