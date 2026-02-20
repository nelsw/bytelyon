package manager

import (
	"context"
	"sync"
	"time"

	"github.com/nelsw/bytelyon/internal/db"
	"github.com/nelsw/bytelyon/internal/model"
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
		time.Sleep(time.Second * 15)
	}
}

func (m *Manager) work() {

	bots, err := db.Builder[model.Bot]().Where("frequency > ?", 0).Find(context.Background())
	if err != nil {
		log.Err(err).Msg("failed to find bots")
		return
	}

	log.Debug().Msgf("bots found [%d]", len(bots))
	if len(bots) == 0 {
		return
	}

	var ready int
	for _, job := range bots {
		if job.UpdatedAt.Add(job.Frequency).Before(time.Now()) {
			ready++
		}
	}

	log.Debug().Msgf("bots ready [%d]", ready)
	if ready == 0 {
		return
	}

	var wg sync.WaitGroup
	for _, bot := range bots {
		wg.Go(func() {
			switch bot.Type {
			case model.NewsBotType:
				news.New(&bot).Work()
			case model.SitemapBotType:
				sitemap.New(&bot).Work()
			case model.SearchBotType:
				search.New(&bot).Work()
			default:
				log.Warn().Msg("unknown bot type")
				return
			}
			// A frequency of 1 means the bot runs once
			// Reset it to 0 so it doesn't run again
			if bot.Frequency == 1 {
				_, err = db.Builder[model.Bot]().
					Where("id = ?", bot.ID).
					Update(context.Background(), "frequency", 0)

				if err != nil {
					log.Err(err).Msg("manager failed to zero out bot frequency")
				}
			}
		})
	}
	wg.Wait()
	log.Debug().Msg("bots deployed")
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
