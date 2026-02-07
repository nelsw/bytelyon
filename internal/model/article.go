package model

import (
	"time"

	"gorm.io/gorm"
)

type Article struct {
	gorm.Model
	JobID     uint
	Job       Job
	URL       string `gorm:"index:idx_article_url,unique"`
	Title     string
	Source    string
	Published time.Time
}
