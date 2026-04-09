package s3

import (
	"context"
	"fmt"
	"strings"

	"github.com/nelsw/bytelyon/pkg/aws"
	"github.com/nelsw/bytelyon/pkg/client"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/oklog/ulid/v2"
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
	if err := client.PutPublicImage(ctx, c, public, k, d); err != nil {
		return "", err
	}
	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", public, k), nil
}

func PutPublicBotResultData(r *model.BotResult, k string, d []byte) (string, error) {
	k = kee(r.UserID, r.Type, r.Target, k)
	if err := client.PutPublicImage(ctx, c, public, k, d); err != nil {
		return "", err
	}
	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", public, k), nil
}

func key(bot *model.Bot, key string) string {
	key = fmt.Sprintf(
		"users/%s/bots/%s/%s/%s",
		bot.UserID,
		bot.Type,
		bot.Target,
		key,
	)
	key = strings.ReplaceAll(key, " ", "-")
	return key
}

func kee(userID ulid.ULID, botType model.BotType, target, key string) string {
	key = fmt.Sprintf(
		"users/%s/bots/%s/%s/%s",
		userID,
		botType,
		target,
		key,
	)
	key = strings.ReplaceAll(key, " ", "-")
	return key
}
