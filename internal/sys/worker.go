package sys

import (
	"errors"
	"sync"
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

type Worker struct {
	pwc  *playwright.Playwright
	uid  ulid.ULID
	busy bool
	stop bool
}

// Work bot jobs for a given user
func (w *Worker) Work() {

	l := log.With().Stringer("uid", w.uid).Logger()

	bots := bot.AllReady(w.uid)
	l.Info().Msgf("bots to work: %d", len(bots))

	if len(bots) == 0 {
		w.Sleep()
		return
	}

	w.busy = true
	var wg sync.WaitGroup
	for _, b := range bot.AllReady(w.uid) {
		wg.Go(func() { Execute(b, w.pwc, w.uid) })
	}
	wg.Wait()
	w.busy = false
	w.Sleep()
}

func (w *Worker) Sleep() {
	for i := 0; i < 10; i++ {
		if time.Sleep(time.Second); w.stop {
			return
		}
	}
	w.Work()
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

	l.Info().Msg("working ...")

	ctx, err := pw.Open(p, b.Headless, fingerprint.Find(u, b.Type, b.Target))
	defer func(x playwright.BrowserContext) {
		pw.Close(x)
		l.Err(err).Msg("... worked!")
	}(ctx)

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
}
