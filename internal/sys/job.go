package sys

import (
	"time"

	"github.com/nelsw/bytelyon/pkg/bot"
	"github.com/nelsw/bytelyon/pkg/news"
	"github.com/nelsw/bytelyon/pkg/pw"
	"github.com/nelsw/bytelyon/pkg/search"
	"github.com/nelsw/bytelyon/pkg/sitemap"
	"github.com/oklog/ulid/v2"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

type Jobs []*Job

type Job struct {
	*bot.Model
	pwc *playwright.Playwright
	uid ulid.ULID
}

func NewJob(
	bot *bot.Model,
	pwc *playwright.Playwright,
	uid ulid.ULID,
) *Job {
	return &Job{
		Model: bot,
		pwc:   pwc,
		uid:   uid,
	}
}

func (j *Job) Work() {

	l := log.With().
		Stringer("uid", j.uid).
		Str("target", j.Target).
		Stringer("type", j.Type).
		Logger()

	l.Trace().Send()

	ctx, err := pw.Open(j.pwc, j.Headless, j.Fingerprint)
	if err != nil {
		l.Err(err).Send()
		return
	}
	defer pw.Close(ctx)

	switch j.Type {
	case bot.News:
		news.Work(ctx, j.uid, j.Target, j.Blacklist, j.RanAt)
	case bot.Search:
		search.Work(ctx, j.uid, j.Target, j.Blacklist)
	case bot.Sitemap:
		sitemap.Work(ctx, j.uid, j.Target)
	}

	// update bot worked at to now
	j.RanAt = time.Now().UTC()

	// reset frequency if set to 1ns (once & stop)
	if j.Frequency == 1 {
		j.Frequency = 0
	}

	// update the storage state of the bot
	var state *playwright.StorageState
	if state, err = ctx.StorageState(); err == nil && state != nil {
		j.Fingerprint = state
	}

	// save bot
	log.Err(bot.Update(j.uid, j.Model)).Send()
}
