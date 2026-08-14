package client

import (
	"testing"

	"github.com/nelsw/bytelyon/apps/mux/internal/logger"
	"github.com/nelsw/bytelyon/apps/mux/internal/model"
	"github.com/rs/zerolog/log"
)

func init() {
	logger.Init()
}

func TestGet(t *testing.T) {

	t.Setenv("WEB_URL", "http://localhost/api")
	t.Setenv("WEB_KEY", "my-random-32-character-x-api-key")

	var bots model.Bots
	if err := Get[model.Bots](); err != nil {
		t.Errorf("failed to get bots: %v", err)
	}
	if len(bots) == 0 || bots == nil {
		t.Errorf("expected at least one bot, got none")
	}
	for _, bot := range bots {
		log.Debug().EmbedObject(bot).Send()
		if bot.ID == 0 {
			t.Errorf("bot ID should not be zero")
		}

	}
}
