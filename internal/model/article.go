package model

import (
	"time"

	"gorm.io/gorm"
)

type Article struct {
	gorm.Model
	JobID     uint
	Job       Job
	URL       string
	Title     string
	Source    string
	Published time.Time
}
