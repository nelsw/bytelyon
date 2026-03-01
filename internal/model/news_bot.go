package model

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	. "github.com/nelsw/bytelyon/internal/config"
)

type NewsBot struct {
	Bot
}

func (b *NewsBot) Desc() *dynamodb.CreateTableInput {
	d := b.Bot.desc()
	d.TableName = TableName(b)
	return d
}
