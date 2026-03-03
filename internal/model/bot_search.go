package model

import (
	"fmt"

	"github.com/google/uuid"
)

type BotSearch struct {
	Bot
	Headless bool        `json:"headless" dynamodbav:"Headless,boolean"`
	State    BroCtxState `json:"state" dynamodbav:"State,boolean"`
}

func (b BotSearch) PageDataPath(id uuid.UUID, idx int, ext string) string {
	return fmt.Sprintf("users/%s/bots/search/%s/%s/%d.%s",
		b.Bot.UserID,
		b.Bot.Target,
		id,
		idx,
		ext,
	)
}
