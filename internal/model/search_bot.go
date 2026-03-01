package model

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	. "github.com/nelsw/bytelyon/internal/config"
)

type SearchBot struct {
	Bot
	Headless bool                `json:"headless" dynamodbav:"Headless,boolean"`
	State    BrowserContextState `json:"state" dynamodbav:"State,boolean"`
}

func (b *SearchBot) Desc() *dynamodb.CreateTableInput {
	d := b.Bot.desc()
	d.TableName = TableName(b)
	return d
}
