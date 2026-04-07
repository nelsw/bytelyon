package manager

import (
	"context"
	"time"

	"github.com/nelsw/bytelyon/pkg/client"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/pw"
	"github.com/nelsw/bytelyon/pkg/repo"
	"github.com/oklog/ulid/v2"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

var (
	users = []*model.User{
		//{ID: ulid.MustParse("01KM01JC9PS1R4X4FDJNFAR4AZ"), Name: "Guest"},
		{ID: ulid.MustParse("01KMXGBJJE2GMCA1A9EXDGF4AJ"), Name: "Stu"},
		{ID: ulid.MustParse("01KM010XK0HY8HWWFPJTZGRF0F"), Name: "Carl"},
	}
)

type Manager struct {
	headless,
	headed playwright.Browser
	stop,
	done bool
}

func (m *Manager) Start() {

	log.Info().Msg("bot manager looking for work")

	for !m.stop {

		log.Info().Msg("bot manager working ...")

		m.done = false
		m.work()
		m.done = true

		log.Info().Msg("bot manager work complete")

		if m.stop {
			return
		}

		d := time.Duration(15) * time.Second

		log.Debug().
			Stringer("duration", d).
			Msg("bot manager sleeping ...")

		time.Sleep(d)
	}
}

func (m *Manager) Stop(ctx context.Context) error {

	m.stop = true

	timer := time.NewTimer(time.Second)

	defer func() {
		timer.Stop()
		log.Debug().Msg("bots manager stopped")
	}()

	log.Info().Msg("bots manager stopping")
	for {
		if m.done {
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

func (m *Manager) work() {

	var err error
	m.headless, err = pw.NewBrowser(true)
	if err != nil {
		log.Err(err).Msg("failed to create headless browser")
		return
	}
	defer func(bro playwright.Browser) {
		_ = bro.Close()
	}(m.headless)

	m.headed, err = pw.NewBrowser(false)
	if err != nil {
		log.Err(err).Msg("failed to create headed browser")
		return
	}
	defer func(bro playwright.Browser) {
		_ = bro.Close()
	}(m.headed)

	var jobs []*Job
	for _, user := range users {
		for _, bot := range repo.FindBots(user.ID) {
			if !bot.IsReady() {
				continue
			}
			var ctx playwright.BrowserContext
			if bot.Fingerprint == nil {
				log.Debug().Msg("fingerprint is nil")
				bot.Fingerprint = model.NewFingerprint()
			}
			if state := bot.Fingerprint.GetState(); bot.Headless {
				ctx, err = client.NewContext(m.headless, state)
			} else {
				ctx, err = client.NewContext(m.headed, state)
			}

			if err != nil {
				log.Err(err).Msgf("failed to create browser context for %s", bot.Target)
				continue
			}
			jobs = append(jobs, &Job{ctx, bot})
		}
	}
	log.Info().Msgf("jobs found: %d", len(jobs))

	//var wg sync.WaitGroup
	for _, j := range jobs {
		//wg.Go(func() { j.Work() })
		j.Work()
	}
	//log.Info().Msg("working jobs")
	//wg.Wait()
	log.Info().Msg("worked jobs")
}

func New() *Manager {
	return new(Manager)
}
