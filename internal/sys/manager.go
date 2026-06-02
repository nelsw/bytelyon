package sys

import (
	"context"
	"math/rand"
	"time"

	"github.com/nelsw/bytelyon/pkg/user"
	"github.com/playwright-community/playwright-go"
)

const shutdownPollIntervalMax = 500 * time.Millisecond

type Manager struct {
	pwc     *playwright.Playwright
	workers []*Worker
	stop    bool
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
	if m.stop {
		return
	}
	for _, w := range m.workers {
		go w.Work()
	}
	time.AfterFunc(10*time.Second, m.Start)
}

func (m *Manager) Stop(ctx context.Context) error {

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

	done := func() bool {
		for _, w := range m.workers {
			if w.busy {
				return false
			}
		}
		return true
	}

	timer := time.NewTimer(nextPollInterval())
	defer timer.Stop()
	for {
		if done() {
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
