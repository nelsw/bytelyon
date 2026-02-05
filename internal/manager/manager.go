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

	m.done = false

	var jobs []*model.Job
	if err := m.db.Where("enabled = ?", true).Find(&jobs).Error; err != nil {
		log.Warn().Err(err).Msg("failed to find enabled jobs")
	}

	var wg sync.WaitGroup
	for _, job := range jobs {
		wg.Go(worker.New(m.db, job).Work)
	}

	time.Sleep(time.Minute)

	m.done = true

	if m.stop {
		return
	}

	m.Start()
}

func (m *Manager) Stop() {
	m.stop = true
}

func (m *Manager) Done() bool {
	return m.stop && m.done
}
