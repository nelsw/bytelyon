package sys

import (
	"context"
	"math/rand"
	"time"

	"github.com/nelsw/bytelyon/pkg/id"
	"github.com/playwright-community/playwright-go"
)

const shutdownPollIntervalMax = 500 * time.Millisecond

type Manager struct {
	pwc     *playwright.Playwright
	workers Workers
}

func NewManager() (*Manager, error) {
	pwc, err := playwright.Run()
	if err != nil {
		return nil, err
	}
	// todo - find all users
	return &Manager{
		pwc,
		Workers{
			NewWorker(pwc, id.ParseULID("01KM010XK0HY8HWWFPJTZGRF0F")),
			NewWorker(pwc, id.ParseULID("01KM01JC9PS1R4X4FDJNFAR4AZ")),
			NewWorker(pwc, id.ParseULID("01KMXGBJJE2GMCA1A9EXDGF4AJ")),
		},
	}, nil
}

func (m *Manager) Start() {
	for _, w := range m.workers {
		go w.Work()
	}
}

func (m *Manager) Stop(ctx context.Context) error {

	pollIntervalBase := time.Millisecond
	nextPollInterval := func() time.Duration {
		// Add 10% jitter.
		interval := pollIntervalBase + time.Duration(rand.Intn(int(pollIntervalBase/10)))
		// Double and clamp for next time.
		pollIntervalBase *= 2
		if pollIntervalBase > shutdownPollIntervalMax {
			pollIntervalBase = shutdownPollIntervalMax
		}
		return interval
	}

	ƒ := func() bool {
		for _, w := range m.workers {
			if w.Quit(); !w.Done() {
				return false
			}
		}
		return true
	}

	timer := time.NewTimer(nextPollInterval())
	defer timer.Stop()
	for {
		if ƒ() {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			timer.Reset(nextPollInterval())
		}
	}
}
