package worker

import (
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

func New(pwc *playwright.Playwright, userID ulid.ULID) *Worker {
	return &Worker{pwc, userID, false, true}
}

func (w *Worker) Start() {
	log.Info().Msg("working")
	for !w.stop {
		w.work()
		w.sleep()
	}
}

func (w *Worker) Stop() error {
	w.stop = true
	for !w.done {
		time.Sleep(time.Second)
	}
	return w.pwc.Stop()
}

func (w *Worker) sleep() {
	logs.PrintNyanCat()
	for i := 0; i < 60 && !w.stop; i++ {
		time.Sleep(time.Second)
	}
}

func (w *Worker) work() {

	bro, err := pw.NewBrowser(w.pwc, true)
	if err != nil {
		log.Err(err).Msgf("failed to create browser for %s", w.userID)
		return
	}

	w.done = false
	defer func() {
		bro.Close()
		w.done = true
	}()

	var jobs []*Job
	for _, bot := range repo.FindBots(w.userID) {

		l := log.With().
			Stringer("type", bot.Type).
			Str("target", bot.Target).
			Logger()

		l.Info().Bool("ready", bot.IsReady()).Send()

		if !bot.IsReady() {
			continue
		}

		if bot.Fingerprint == nil {
			bot.Fingerprint = model.NewFingerprint()
		}

		var ctx playwright.BrowserContext
		if ctx, err = pw.NewBrowserContext(bro, bot.Fingerprint.GetState()); err != nil {
			l.Err(err).Msg("failed to create browser context")
			continue
		}

		jobs = append(jobs, &Job{ctx, bot})
	}

	for _, job := range jobs {
		job.Work()
	}
}
