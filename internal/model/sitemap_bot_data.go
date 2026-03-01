package model

import "github.com/google/uuid"

type SitemapBotData struct {
	BotID    uuid.UUID `json:"botID" dynamodbav:"BotID,binary"`
	URL      string    `json:"url" dynamodbav:"URL,string"`
	Domain   string    `json:"domain" dynamodbav:"Domain,string"`
	Relative []string  `json:"relative" dynamodbav:"Relative,stringset"`
	Remote   []string  `json:"remote" dynamodbav:"Remote,stringset"`
}
