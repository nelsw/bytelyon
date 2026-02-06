package model

import (
	"time"

	"gorm.io/gorm"
)

type Job struct {
	gorm.Model
	Enabled   bool
	Type      JobType
	Frequency time.Duration
	Target    string   `gorm:"type:varchar(255)"`
	BlackList []string `gorm:"serializer:json"`
}

func (j Job) Ignore() map[string]bool {
	m := map[string]bool{}
	for _, s := range j.BlackList {
		m[s] = true
	}
	return m
}
