package worker

import (
	"time"

	"github.com/nelsw/bytelyon/pkg/logs"
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

		w.done = false
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
