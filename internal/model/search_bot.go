package model

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	. "github.com/nelsw/bytelyon/internal/config"
	. "github.com/nelsw/bytelyon/internal/util"
)

type SearchBot struct {
	Bot
	Headless bool `json:"headless" dynamodbav:"Headless,boolean"`
}

func (b *SearchBot) Name() string {
	return "ByteLyon_" + ModeTitle() + "_Search_Bot"
}

func (b *SearchBot) Desc() *dynamodb.CreateTableInput {
	d := b.Bot.desc()
	d.TableName = Ptr(b.Name())
	return d
}
