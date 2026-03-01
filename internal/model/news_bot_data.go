package model

import (
	"time"

	"github.com/google/uuid"
)

type NewsBotData struct {
	BotID       uuid.UUID `json:"botID" dynamodbav:"BotID,binary"`
	URL         string    `json:"url" dynamodbav:"URL,binary"`
	Title       string    `json:"title" dynamodbav:"Title,string"`
	Source      string    `json:"source" dynamodbav:"Source,string"`
	Description string    `json:"description" dynamodbav:"Description,string"`
	Published   time.Time `json:"published" dynamodbav:"Published,number"`
}
