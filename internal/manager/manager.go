package manager

import (
	"sync"
	"time"

	"github.com/nelsw/bytelyon/internal/model"
	"github.com/nelsw/bytelyon/internal/worker"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type Manager struct {
	db   *gorm.DB
	stop bool
	done bool
}

func New(db *gorm.DB) *Manager {
	return &Manager{db: db}
}

func (m *Manager) Start() {

	if m.stop {
		return
	}

	m.done = false

	var jobs []*model.Job
	if err := m.db.Where("enabled = ?", true).Find(&jobs).Error; err != nil {
		log.Warn().Err(err).Msg("failed to find enabled jobs")
	}

	var wg sync.WaitGroup
	for _, job := range jobs {
		wg.Go(worker.New(m.db, job).Work)
	}

	if m.done = true; m.stop {
		return
	}

	time.Sleep(time.Minute)

	m.Start()
}

func (m *Manager) Stop() bool {
	m.stop = true
	if !m.done {
		time.Sleep(time.Second)
	}
	return m.Stop()
}
