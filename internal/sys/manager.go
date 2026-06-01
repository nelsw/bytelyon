package sys

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/nelsw/bytelyon/pkg/user"
	"github.com/playwright-community/playwright-go"
)

const shutdownPollIntervalMax = 500 * time.Millisecond

type Manager struct {
	pwc     *playwright.Playwright
	workers []*Worker
	stop    bool
	done    bool
}

func NewManager() (m *Manager, err error) {

	m = new(Manager)

	if m.pwc, err = playwright.Run(); err != nil {
		return nil, err
	}

	for _, uid := range user.IDs() {
		m.workers = append(m.workers, NewWorker(m.pwc, uid))
	}

	return
}

func (m *Manager) Start() {

	m.done = false
	var wg sync.WaitGroup
	for _, w := range m.workers {
		wg.Go(w.Work)
	}
	wg.Wait()
	m.done = true

	for i := 0; i < 10; i++ {
		if time.Sleep(time.Second); m.stop {
			return
		}
	}

	m.Start()
}

func (m *Manager) Stop(ctx context.Context) error {

	m.stop = true
	for _, w := range m.workers {
		w.stop = true
	}

	pollIntervalBase := time.Millisecond
	nextPollInterval := func() time.Duration {
		// Add 10% jitter.
		interval := pollIntervalBase + time.Duration(rand.Intn(int(pollIntervalBase/10)))
		// Double and clamp for next time.
		if pollIntervalBase *= 2; pollIntervalBase > shutdownPollIntervalMax {
			pollIntervalBase = shutdownPollIntervalMax
		}
		return interval
	}

	timer := time.NewTimer(nextPollInterval())
	defer timer.Stop()
	for {
		if m.done {
			return m.pwc.Stop()
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			timer.Reset(nextPollInterval())
		}
	}
}
