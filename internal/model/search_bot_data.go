package model

import "github.com/google/uuid"

type SearchBotData struct {
	BotID uuid.UUID `json:"botID" dynamodbav:"BotID,binary"`
	// page uri set?
	// serp result data?
}
