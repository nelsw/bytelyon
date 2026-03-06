package model

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	. "github.com/nelsw/bytelyon/internal/util"
	"github.com/oklog/ulid/v2"
)

type BotSitemapResult struct {
	Model
	ID       ulid.ULID `json:"ID" dynamodbav:"ID,binary"`
	Target   string    `json:"target" dynamodbav:"Target,string"`
	Relative []string  `json:"relative" dynamodbav:"Relative,stringset"`
	Remote   []string  `json:"remote" dynamodbav:"Remote,stringset"`
}
