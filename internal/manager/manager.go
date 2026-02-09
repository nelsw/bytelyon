package manager

import (
	"sync"
	"time"

	"github.com/nelsw/bytelyon/internal/model"
	"github.com/nelsw/bytelyon/internal/worker/news"
	"github.com/nelsw/bytelyon/internal/worker/search"
	"github.com/nelsw/bytelyon/internal/worker/sitemap"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type Manager struct {
	*gorm.DB
	stop, done bool
}

func New(db *gorm.DB) *Manager {
	return &Manager{DB: db}
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

	var bots []*model.Bot
	if err := m.Where("frequency > ?", 0).Find(&bots).Error; err != nil {
		log.Panic().Err(err).Send()
	}

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
				news.New(m.DB, bot).Work()
			case model.SitemapBotType:
				sitemap.New(m.DB, bot).Work()
			case model.SearchBotType:
				search.New(m.DB, bot).Work()
			default:
				log.Warn().Msg("unknown bot type")
				return
			}
			// A frequency of 1 means the bot runs once
			// Reset it to 0 so it doesn't run again
			if bot.Frequency == 1 {
				bot.Frequency = 0
			}
			m.Save(&bot)
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
