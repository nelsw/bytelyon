package model

import (
	"time"

	"github.com/google/uuid"
)

type Bot struct {
	UserID    uuid.UUID     `json:"userID" dynamodbav:"UserID,binary"`
	BotID     uuid.UUID     `json:"botID" dynamodbav:"BotID,binary"`
	Target    string        `json:"target" dynamodbav:"Target"`
	Frequency time.Duration `json:"frequency" dynamodbav:"Frequency,number"`
	BlackList []string      `json:"blackList" dynamodbav:"BlackList,stringset"`
	UpdatedAt time.Time     `json:"updatedAt" dynamodbav:"UpdatedAt,number"`
}

func (b *Bot) Ignore() map[string]bool {
	m := map[string]bool{}
	for _, s := range b.BlackList {
		m[s] = true
	}
	return m
}
