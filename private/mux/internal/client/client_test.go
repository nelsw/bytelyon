package client

import (
	"testing"

	"github.com/nelsw/bytelyon/apps/mux/internal/logger"
	"github.com/nelsw/bytelyon/apps/mux/internal/model"
	"github.com/stretchr/testify/assert"
)

func init() {
	logger.Init()
}

func TestGet(t *testing.T) {

	t.Setenv("WEB_URL", "http://localhost/api")
	t.Setenv("WEB_KEY", "my-random-32-character-x-api-key")

	var bots model.Bots
	if err := Get[model.Bots](); err != nil {
		t.Fatalf("failed to get bots: %v", err)
	}
	assert.NotEmpty(t, bots)
	assert.NotNil(t, bots)
}
