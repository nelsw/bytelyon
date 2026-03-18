package model

import (
	"fmt"

	"github.com/oklog/ulid/v2"
)

type BotSearch struct {
	Bot
	Headless bool        `json:"headless" dynamodbav:"Headless,boolean"`
	State    BroCtxState `json:"state" dynamodbav:"Fingerprint,boolean"`
}

func (b BotSearch) PageDataPath(id ulid.ULID, idx int, ext string) string {
	return fmt.Sprintf("users/%s/bots/search/%s/%s/%d.%s",
		b.Bot.UserID,
		b.Bot.Target,
		id,
		idx,
		ext,
	)
}
