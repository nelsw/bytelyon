package sys

import (
	"errors"
	"time"

	"github.com/nelsw/bytelyon/pkg/bot"
	"github.com/nelsw/bytelyon/pkg/fingerprint"
	"github.com/nelsw/bytelyon/pkg/news"
	"github.com/nelsw/bytelyon/pkg/pw"
	"github.com/nelsw/bytelyon/pkg/search"
	"github.com/nelsw/bytelyon/pkg/sitemap"
	"github.com/oklog/ulid/v2"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

type Job struct {
	*bot.Model
	pwc *playwright.Playwright
	uid ulid.ULID
	err error
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

func Execute(
	b *bot.Model,
	p *playwright.Playwright,
	u ulid.ULID,
) {

	l := log.With().
		Stringer("uid", u).
		Str("target", b.Target).
		Stringer("type", b.Type).
		Logger()

	l.Info().Msg("executing ...")

	ctx, err := pw.Open(p, b.Headless, fingerprint.Find(u, b.Type, b.Target))
	defer func() {
		pw.Close(ctx)
		l.Err(err).Msg("executed")
	}()

	if err != nil {
		return
	}

	switch b.Type {
	case bot.News:
		news.Work(ctx, u, b.Target, b.Blacklist, b.RanAt)
	case bot.Search:
		search.Work(ctx, u, b.Target, b.Blacklist)
	case bot.Sitemap:
		sitemap.Work(ctx, u, b.Target)
	}

	// update when this bot was run, and when to run next
	if b.RanAt = time.Now().UTC(); b.Frequency == 1 {
		// reset frequency if set to 1ns (once & stop)
		b.Frequency = 0
	}

	// update bot
	err = errors.Join(err, bot.Update(u, b))

	// update the storage state of the bot
	if state, _ := ctx.StorageState(); state != nil {
		err = errors.Join(err, fingerprint.Save(u, b.Type, b.Target, state))
	}

	return
}

func (j *Job) AddError(err error) {
	j.err = errors.Join(j.err, err)
}

func (j *Job) Work() {

	l := log.With().
		Str("ƒ", "job").
		Stringer("uid", j.uid).
		Str("target", j.Target).
		Stringer("type", j.Type).
		Logger()

	l.Trace().Send()

	ctx, err := pw.Open(j.pwc, j.Headless, fingerprint.Find(j.uid, j.Type, j.Target))
	if err != nil {
		j.AddError(err)
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

	// update when this bot was run, and when to run next
	if j.RanAt = time.Now().UTC(); j.Frequency == 1 {
		// reset frequency if set to 1ns (once & stop)
		j.Frequency = 0
	}

	// update bot
	if err = bot.Update(j.uid, j.Model); err != nil {
		j.AddError(err)
	}

	// update the storage state of the bot
	var state *playwright.StorageState
	if state, err = ctx.StorageState(); err != nil {
		j.AddError(err)
	} else if err = fingerprint.Save(j.uid, j.Type, j.Target, state); err != nil {
		j.AddError(err)
	}
}
