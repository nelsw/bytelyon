package job

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/nelsw/bytelyon/pkg/db"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/rs/zerolog/log"
)

type Job struct {
	ctx   context.Context
	db    *dynamodb.Client
	s3    *s3.Client
	bot   *model.Bot
	rules map[string]bool
}

func New(ctx context.Context, db *dynamodb.Client, s3 *s3.Client, bot *model.Bot) *Job {
	rules := make(map[string]bool)
	for _, s := range bot.BlackList {
		rules[s] = false
	}
	return &Job{ctx, db, s3, bot, rules}
}

func (j *Job) Work() {
	switch j.bot.Type {
	case "search":
		j.doSearch()
	case "sitemap":
		j.doSitemap()
	case "news":
		j.doNews()
	default:
		log.Warn().Msgf("bot type [%s] not supported", j.bot.Type)
		return
	}

	// update bot worked at to now
	j.bot.WorkedAt = time.Now().UTC()

	// reset frequency if set to 1ns (once & stop)
	if j.bot.Frequency == 1 {
		j.bot.Frequency = 0
	}

	// save bot
	if err := db.PutItem(j.bot); err != nil {
		log.Warn().Err(err).Msg("Failed to Save Search Bot (DB)")
	}
}
