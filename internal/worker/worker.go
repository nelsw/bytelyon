package worker

import (
	"context"
	"time"

	"github.com/nelsw/bytelyon/internal/pw"
	"github.com/nelsw/bytelyon/pkg/logs"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/repo"
	"github.com/oklog/ulid/v2"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

type Worker struct {
	pwc        *playwright.Playwright
	userID     ulid.ULID
	stop, done bool
}

func New(userID ulid.ULID) *Worker {
	return &Worker{
		pwc:    pw.Run(),
		userID: userID,
		stop:   false,
		done:   true,
	}
}

func (w *Worker) Start() {
	log.Info().Msg("working")
	for !w.stop {
		w.work()
		w.sleep()
	}
}

func (w *Worker) Stop(ctx context.Context) error {

	w.stop = true

	timer := time.NewTimer(time.Second)
	defer func() {
		timer.Stop()
	}()

	w.pwc.Stop()

	for {
		if w.done {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			timer.Reset(time.Second)
		}
	}
}

func (w *Worker) sleep() {
	logs.PrintNyanCat()
	for i := 0; i < 60 && !w.stop; i++ {
		time.Sleep(time.Second)
	}
}

func (w *Worker) work() {

	w.done = false
	defer func() {
		w.done = true
	}()

	bro, err := pw.NewBrowser(w.pwc, true)
	if err != nil {
		log.Err(err).Msgf("failed to create browser for %s", w.userID)
		return
	}
	defer bro.Close()

	ƒ := func(bot *model.Bot) {

		l := log.With().
			Stringer("type", bot.Type).
			Str("target", bot.Target).
			Logger()

		if l.Info().Bool("ready", bot.IsReady()).Send(); !bot.IsReady() {
			return
		}

		if bot.Fingerprint == nil {
			bot.Fingerprint = model.NewFingerprint()
		}

		var ctx playwright.BrowserContext
		if ctx, err = pw.NewBrowserContext(bro, bot.Fingerprint.GetState()); err != nil {
			l.Err(err).Msg("failed to create browser context")
			return
		}
		defer ctx.Close()

		j := &Job{ctx, bot}
		j.Work()
	}

	for _, bot := range repo.FindBots(w.userID) {
		ƒ(bot)
	}
}
