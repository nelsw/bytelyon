package model

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	. "github.com/nelsw/bytelyon/internal/config"
)

type SitemapBot struct {
	Bot
}

func (b *SitemapBot) Validate() error {
	if ok := strings.HasPrefix(b.Target, "https://"); !ok {
		return fmt.Errorf("bad url, must begin with https://")
	}
	return nil
}

func (b *SitemapBot) Desc() *dynamodb.CreateTableInput {
	d := b.Bot.desc()
	d.TableName = TableName(b)
	return d
}
