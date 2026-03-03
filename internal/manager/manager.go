package manager

import (
	"context"
	"sync"
	"time"

	"github.com/nelsw/bytelyon/internal/model"
	"github.com/nelsw/bytelyon/internal/service/db"
	"github.com/nelsw/bytelyon/internal/worker/news"
	"github.com/nelsw/bytelyon/internal/worker/search"
	"github.com/nelsw/bytelyon/internal/worker/sitemap"
	"github.com/rs/zerolog/log"
)

type Manager struct {
	stop, done bool
}

func New() *Manager {
	return &Manager{}
}

func (m *Manager) Start() {

	log.Info().Msg("bot manager started")

	for !m.stop {
		log.Debug().Msg("bot manager working")
		m.done = false
		m.work()

		if m.done = true; m.stop {
			return
		}

		log.Debug().Msg("bot manager sleeping")
		time.Sleep(15 * time.Second)
	}
}

func (m *Manager) work() {

	users, err := db.Scan[model.User](model.User{})
	if err != nil {
		log.Error().Err(err).Msg("user scan failed")
		return
	}

	log.Debug().Int("size", len(users)).Msg("users found")

	var wg sync.WaitGroup
	for _, user := range users {
		wg.Go(func() {
			m.workSearchBots(user)
			m.workSitemapBots(user)
			m.workNewsBots(user)
			log.Debug().Msg("robots deployed")
		})
	}
	log.Debug().Msg("waiting for all robots to deploy")
	wg.Wait()
	log.Debug().Msg("all robots deployed")
}

func (m *Manager) Stop(ctx context.Context) error {

	m.stop = true

	timer := time.NewTimer(time.Second)

	defer func() {
		timer.Stop()
		log.Debug().Msg("bot manager stopped")
	}()

	log.Info().Msg("bot manager stopping")
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

func (m *Manager) getUsers() []model.User {
	arr, err := db.Scan[model.User](model.User{})
	if err != nil {
		log.Err(err).Msg("failed to scan users")
		return nil
	}
	return arr
}

func (m *Manager) workSearchBots(user model.User) {

	bots, err := db.Query(model.BotSearch{}, user.ID())
	if err != nil {
		log.Err(err).Msg("failed to query search bots")
		return
	}

	var wg sync.WaitGroup
	for _, b := range bots {
		wg.Go(func() {
			if !b.IsReady() {
				log.Debug().Msgf("search bot is not ready")
			} else {
				log.Debug().Msg("search bot is ready to work")
				search.New(&b).Work()
			}
		})
	}
	wg.Wait()
}

func (m *Manager) workSitemapBots(user model.User) {

	bots, err := db.Query[model.BotSitemap](model.BotSitemap{}, user.ID())
	if err != nil {
		log.Err(err).Msg("failed to query sitemap bots")
		return
	}

	var wg sync.WaitGroup
	for _, b := range bots {
		wg.Go(func() {
			if !b.IsReady() {
				log.Debug().Msgf("sitemap bot is not ready")
			} else {
				log.Debug().Msg("sitemap bot is ready to work")
				sitemap.New(&b).Work()
			}
		})
	}
	wg.Wait()
}

func (m *Manager) workNewsBots(user model.User) {
	bots, err := db.Query[model.BotNews](model.BotNews{}, user.ID())
	if err != nil {
		log.Err(err).Msg("failed to query news bots")
		return
	}

	var wg sync.WaitGroup
	for _, b := range bots {
		wg.Go(func() {
			if !b.IsReady() {
				log.Debug().Msgf("news bot is not ready")
			} else {
				log.Debug().Msg("news bot is ready to work")
				news.New(&b).Work()
			}
		})
	}
	wg.Wait()
}
