package sys

import (
	"github.com/nelsw/bytelyon/pkg/bot"
	"github.com/oklog/ulid/v2"
	"github.com/playwright-community/playwright-go"
)

type Worker struct {
	pwc  *playwright.Playwright
	uid  ulid.ULID
	stop bool
}

func NewWorker(
	pwc *playwright.Playwright,
	uid ulid.ULID,
) *Worker {
	return &Worker{
		pwc: pwc,
		uid: uid,
	}
}

func (w *Worker) Work() {

	if w.stop {
		return
	}

	jobs := make(chan *Job, 3)
	done := make(chan bool)
	go func() {
		for {
			if j, more := <-jobs; more {
				j.Work()
				continue
			}
			done <- true
			break
		}
	}()

	for _, b := range bot.AllReady(w.uid) {
		if w.stop {
			break
		}
		jobs <- NewJob(b, w.pwc, w.uid)
	}
	close(jobs)
	<-done
}
