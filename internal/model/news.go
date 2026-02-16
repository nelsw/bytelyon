package model

import (
	"time"

	"gorm.io/gorm"
)

type News struct {
	gorm.Model
	Bot         *Bot
	BotID       uint
	URL         string `gorm:"index:idx_news_url,unique"`
	Title       string
	Source      string
	Description string
	Published   time.Time
}
