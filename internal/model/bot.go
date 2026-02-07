package model

import (
	"fmt"
	"regexp"
	"time"

	"gorm.io/gorm"
)

var (
	urlValidationRegex = regexp.MustCompile(`https?://(www\.)?[-a-zA-Z0-9@:%._+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_+.~#?&/=]*)`)
)

type Bot struct {
	ID        uint
	Type      BotType        `gorm:"index:idx_job_type_target_deleted,unique"`
	Target    string         `gorm:"index:idx_job_type_target_deleted,unique"`
	DeletedAt gorm.DeletedAt `gorm:"index:idx_job_type_target_deleted,unique"`
	CreatedAt int
	UpdatedAt int
	Frequency time.Duration
	BlackList []string `gorm:"serializer:json"`
}

func (b *Bot) Ignore() map[string]bool {
	m := map[string]bool{}
	for _, s := range b.BlackList {
		m[s] = true
	}
	return m
}

func (b *Bot) Validate() error {
	if b.Type != ArticleBotType {
		if err := urlValidationRegex.MatchString(b.Target); err {
			return fmt.Errorf("bad url, must begin with https://")
		}
	}
	return nil
}

func (b *Bot) ReadyToWork() bool {
	return time.Unix(int64(b.UpdatedAt), 0).Add(b.Frequency).Before(time.Now())
}
