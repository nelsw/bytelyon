package model

import "gorm.io/gorm"

type Search struct {
	gorm.Model
	Bot   Bot
	BotID uint
	Pages []*SearchPage
}
