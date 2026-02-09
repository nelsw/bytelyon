package model

import (
	"time"
)

type News struct {
	*Model
	Bot         *Bot `json:",omitempty"`
	BotID       uint
	URL         string `gorm:"index:idx_news_url,unique"`
	Title       string
	Source      string
	Description string
	Published   time.Time
}
