package model

import (
	"encoding/base64"
	"fmt"
	"time"
)

type BotSearch struct {
	Bot
	Headless bool        `json:"headless" dynamodbav:"Headless,boolean"`
	State    BroCtxState `json:"state" dynamodbav:"State,boolean"`
}

func (b BotSearch) PageDataPath(url, ext string) string {
	return fmt.Sprintf("users/%s/bots/search/%s/%s/%s.%s",
		b.Bot.UserID,
		b.Bot.Target,
		b.UpdatedAt.Truncate(time.Minute),
		base64.URLEncoding.EncodeToString([]byte(url)),
		ext,
	)
}
