package model

import (
	"github.com/oklog/ulid/v2"
)

type BotSitemapResult struct {
	Model
	ID       ulid.ULID `json:"URL" dynamodbav:"URL,binary"`
	Target   string    `json:"target" dynamodbav:"Target,string"`
	Relative []string  `json:"relative" dynamodbav:"Relative,stringset"`
	Remote   []string  `json:"remote" dynamodbav:"Remote,stringset"`
}
