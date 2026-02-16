package model

import "gorm.io/gorm"

type Search struct {
	gorm.Model
	Bot   *Bot
	Pages []*SearchPage
	BotID uint
}
