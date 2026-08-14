package main

import (
	"github.com/nelsw/bytelyon/apps/mux/internal/db"
	"github.com/nelsw/bytelyon/apps/mux/internal/logger"
	"github.com/nelsw/bytelyon/apps/mux/internal/subscriber"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	log.Logger = logger.Make(zerolog.TraceLevel)
}

func main() {
	defer db.Close()

	sub := subscriber.Start()

	if err := db.Pub("bots", "wat"); err != nil {
		log.Error().Err(err).Msg("Failed to publish to Redis")
	}

	sub.Stop()
}
