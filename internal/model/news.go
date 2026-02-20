package model

import (
	"time"

	"gorm.io/gorm"
)

type News struct {
	gorm.Model
	Bot         *Bot
	BotID       uint   `gorm:"index:idx_news_bot_id_url,unique"`
	URL         string `gorm:"index:idx_news_bot_id_url,unique"`
	Title       string
	Source      string
	Description string
	Published   time.Time
}
