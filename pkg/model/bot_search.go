package model

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type BotSearch struct {
	Bot
	Pages []Page
}

func (b BotSearch) PageItems() []types.AttributeValue {
	var items []types.AttributeValue
	for _, p := range b.Pages {
		items = append(items, &types.AttributeValueMemberM{Value: p.Item()})
	}
	return items
}

func (b BotSearch) StoragePath(idx int, ext string) string {
	return fmt.Sprintf("users/%s/bots/search/%s/%s/%d.%s",
		b.Bot.UserID,
		b.Bot.Target,
		b.CreatedAt.Format(time.RFC3339),
		idx,
		ext,
	)
}
