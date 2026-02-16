package manager

import (
	"sync"
	"time"

	"github.com/nelsw/bytelyon/internal/db"
	"github.com/nelsw/bytelyon/internal/model"
	"github.com/nelsw/bytelyon/internal/worker/news"
	"github.com/nelsw/bytelyon/internal/worker/search"
	"github.com/nelsw/bytelyon/internal/worker/sitemap"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type Manager struct {
	stop, done bool
}

func New() *Manager {
	return &Manager{}
}

func (m *Manager) Start() {

	log.Info().Msg("bot manager started")

	if m.stop {
		return
	}

	m.done = false
	m.work()
	m.done = true

	if m.stop {
		return
	}

	log.Info().Msg("bot manager sleeping")
	time.Sleep(time.Minute)
	m.Start()
}

func (m *Manager) work() {

	bots, _ := db.Find[*model.Bot](func(db *gorm.DB) *gorm.DB {
		return db.Where("frequency > ?", 0)
	})

	log.Info().Msgf("bots found [%d]", len(bots))
	if len(bots) == 0 {
		return
	}

	var ready int
	for _, job := range bots {
		if job.ReadyToWork() {
			ready++
		}
	}

	log.Info().Msgf("bots ready [%d]", ready)
	if ready == 0 {
		return
	}

	var wg sync.WaitGroup
	for _, bot := range bots {
		wg.Go(func() {
			switch bot.Type {
			case model.NewsBotType:
				news.New(bot).Work()
			case model.SitemapBotType:
				sitemap.New(bot).Work()
			case model.SearchBotType:
				search.New(bot).Work()
			default:
				log.Warn().Msg("unknown bot type")
				return
			}
			// A frequency of 1 means the bot runs once
			// Reset it to 0 so it doesn't run again
			if bot.Frequency == 1 {
				bot.Frequency = 0
			}
			db.Save(&bot)
		})
	}
	wg.Wait()
	log.Info().Msg("bots deployed")
}

func (m *Manager) Stop() {
	log.Info().Msg("bot manager stopping")
	for m.stop = true; !m.done; {
		time.Sleep(time.Second)
	}
	log.Info().Msg("bot manager stopped")
}
