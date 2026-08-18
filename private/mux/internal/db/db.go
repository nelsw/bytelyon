package db

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

var (
	rdb *redis.Client
)

func init() {
	rdb = redis.NewClient(&redis.Options{Addr: "localhost:6379", DB: 13})
}

func Pub(name string, a any) error {
	log.Debug().Str("channel", name).Msg("Publishing to Redis")
	return rdb.Publish(context.Background(), name, a).Err()
}

func Sub(name string) *redis.PubSub {
	log.Debug().Str("channel", name).Msg("Subscribing to Redis")
	return rdb.Subscribe(context.Background(), name)
}

func Close() {
	if err := rdb.Close(); err != nil {
		log.Warn().Err(err).Msg("Failed to close Redis connection")
	}
}
