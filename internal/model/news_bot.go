package model

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	. "github.com/nelsw/bytelyon/internal/config"
	. "github.com/nelsw/bytelyon/internal/util"
)

type NewsBot struct {
	Bot
}

func (b *NewsBot) Name() string {
	return "ByteLyon_" + ModeTitle() + "_News_Bot"
}

func (b *NewsBot) Desc() *dynamodb.CreateTableInput {
	d := b.Bot.desc()
	d.TableName = Ptr(b.Name())
	return d
}
