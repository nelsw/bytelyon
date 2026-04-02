package s3

import (
	"context"
	"fmt"

	"github.com/nelsw/bytelyon/pkg/aws"
	"github.com/nelsw/bytelyon/pkg/client"
	"github.com/nelsw/bytelyon/pkg/model"
)

const private = "bytelyon-private"
const public = "bytelyon-public"

var ctx = context.Background()
var c = aws.S3()

func PutPrivateBotData(b *model.Bot, k string, d []byte) (string, error) {
	k = key(b, k)
	if err := client.PutObject(ctx, c, private, k, d); err != nil {
		return "", err
	}
	return k, nil
}

func PutPublicBotData(b *model.Bot, k string, d []byte) (string, error) {
	k = key(b, k)
	if err := client.PutObject(ctx, c, public, k, d); err != nil {
		return "", err
	}
	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", public, k), nil
}

func GetPublicObject(key string) ([]byte, error) {
	return client.GetObject(ctx, c, public, key)
}

func key(bot *model.Bot, key string) string {
	return fmt.Sprintf(
		"users/%s/bots/%s/%s/%s",
		bot.UserID,
		bot.Type,
		bot.Target,
		key,
	)
}
