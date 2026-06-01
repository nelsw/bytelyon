package sys

import (
	"github.com/nelsw/bytelyon/pkg/bot"
	"github.com/oklog/ulid/v2"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

type Workers []*Worker
type Worker struct {
	pwc        *playwright.Playwright
	uid        ulid.ULID
	quit, done bool
}

func NewWorker(pwc *playwright.Playwright, uid ulid.ULID) *Worker {
	return &Worker{pwc: pwc, uid: uid}
}

func (w *Worker) Quit() {
	w.quit = true
}

func (w *Worker) Done() bool {
	return w.done
}

func (w *Worker) Work() {

	var jobs Jobs
	for _, t := range bot.Types {
		for _, b := range bot.FindAll(w.uid, t) {
			if b.IsReady() {
				jobs = append(jobs, NewJob(b, w.pwc, w.uid))
			}
		}
	}

	log.Debug().
		Int("jobs", len(jobs)).
		Stringer("uid", w.uid).
		Send()

	queue := make(chan *Job, 3)
	done := make(chan bool)

	go func() {
		for {
			if q, more := <-queue; more {
				q.Work()
				continue
			}
			done <- true
			break
		}
	}()

	for _, j := range jobs {
		if w.quit {
			break
		}
		queue <- j
	}
	close(queue)
	<-done
	w.done = true
}
