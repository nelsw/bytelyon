package worker

import (
	"time"

	"github.com/nelsw/bytelyon/pkg/bot"
	"github.com/nelsw/bytelyon/pkg/logs"
	"github.com/nelsw/bytelyon/pkg/pw"
	"github.com/oklog/ulid/v2"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

type Worker struct {
	pwc *playwright.Playwright
	uid ulid.ULID
	stop,
	done bool
}

func New(pwc *playwright.Playwright, userID ulid.ULID) *Worker {
	return &Worker{pwc, userID, false, true}
}

func (w *Worker) Start() {
	for !w.stop {

		log.Info().Msg("working")

		var bro playwright.Browser
		var ctx playwright.BrowserContext
		var err error

		w.done = false
		for _, typ := range []string{"news", "search", "sitemap"} {
			for _, b := range bot.Find(w.uid, typ) {

				l := log.With().
					Str("type", typ).
					Str("target", b.Target).
					Logger()

				l.Trace().Send()

				if !b.IsReady() {
					l.Debug().Msgf("bot %s is not ready", w.uid)
				} else if bro, err = pw.NewBrowser(w.pwc, b.Headless); err != nil {
					l.Err(err).Msgf("failed to create browser for %s", w.uid)
				} else if ctx, err = pw.NewBrowserContext(bro, b.Fingerprint.GetState()); err != nil {
					l.Err(err).Msg("failed to create browser context")
				} else {
					bot.Run(bro, ctx, b, w.uid)
				}
			}
		}
		w.done = true

		logs.PrintNyanCat()
		for i := 0; i < 60 && !w.stop; i++ {
			time.Sleep(time.Second)
		}
	}
}

func (w *Worker) Stop() error {
	w.stop = true
	for !w.done {
		time.Sleep(time.Second)
	}
	return w.pwc.Stop()
}
