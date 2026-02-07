package model

import (
	"time"

	"gorm.io/gorm"
)

type Article struct {
	gorm.Model
	Bot       Bot
	BotID     uint
	URL       string `gorm:"index:idx_article_url,unique"`
	Title     string
	Source    string
	Published time.Time
}
