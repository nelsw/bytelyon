package model

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type SearchBot struct {
	Bot
	Headless bool                `json:"headless" dynamodbav:"Headless,boolean"`
	State    BrowserContextState `json:"state" dynamodbav:"State,boolean"`
}

func (b SearchBot) Desc() dynamodb.CreateTableInput { return b.Bot.desc() }

func (b SearchBot) PageDataPath(url, ext string) string {
	return fmt.Sprintf("users/%s/bots/search/%s/%s/%s.%s",
		b.Bot.UserID,
		b.Bot.Target,
		b.UpdatedAt.Truncate(time.Minute),
		base64.URLEncoding.EncodeToString([]byte(url)),
		ext,
	)
}
