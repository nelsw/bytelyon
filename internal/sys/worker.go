package sys

import (
	"github.com/nelsw/bytelyon/pkg/bot"
	"github.com/oklog/ulid/v2"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

type Worker struct {
	pwc  *playwright.Playwright
	uid  ulid.ULID
	stop bool
	busy bool
}

func NewWorker(pwc *playwright.Playwright, uid ulid.ULID) *Worker {
	return &Worker{
		pwc: pwc,
		uid: uid,
	}
}

// Work runs bot jobs for a user
func (w *Worker) Work() {

	if w.busy || w.stop {
		return
	}
	w.busy = true

	log.Info().Stringer("uid", w.uid).Msg("working ...")

	jobs := make(chan *bot.Model, 3)
	done := make(chan bool, 1)

	go func() {
		for {
			if j, more := <-jobs; more {
				go Execute(j, w.pwc, w.uid)
				continue
			}
			done <- true
			break
		}
	}()

	for _, job := range bot.AllReady(w.uid) {
		if w.stop {
			log.Info().Stringer("uid", w.uid).Msg("stopping ...")
			break
		}
		jobs <- job
	}

	close(jobs)
	<-done
	w.busy = false

	log.Info().Stringer("uid", w.uid).Msg("worked")
}
