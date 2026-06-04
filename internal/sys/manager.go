package sys

import (
	"context"
	"math/rand"
	"time"

	"github.com/nelsw/bytelyon/pkg/user"
	"github.com/oklog/ulid/v2"
	"github.com/playwright-community/playwright-go"
)

const shutdownPollIntervalMax = 500 * time.Millisecond

type Manager struct {
	pwc     *playwright.Playwright
	workers map[ulid.ULID]*Worker
	stop    bool
}

func NewManager() (*Manager, error) {
	pwc, err := playwright.Run()
	if err != nil {
		return nil, err
	}
	return &Manager{
		pwc:     pwc,
		workers: make(map[ulid.ULID]*Worker),
	}, nil
}

func (m *Manager) Start() {

	if m.stop {
		return
	}

	for _, uid := range user.IDs() {
		if _, ok := m.workers[uid]; ok {
			continue
		}
		m.workers[uid] = &Worker{
			pwc: m.pwc,
			uid: uid,
		}
		go m.workers[uid].Work()
	}

	time.AfterFunc(time.Minute, m.Start)
}

func (m *Manager) Stop(ctx context.Context) error {

	m.stop = true
	done := make(chan bool)
	go func() {
		for _, w := range m.workers {
			if w.stop = true; w.busy {
				return
			}
		}
		done <- true
	}()

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
		select {
		case <-done:
			return m.pwc.Stop()
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			timer.Reset(nextPollInterval())
		}
	}
}
