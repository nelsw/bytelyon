package subscriber

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/nelsw/bytelyon/apps/mux/internal/logger"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func TestDoBot(t *testing.T) {

	_ = os.Setenv("APP_ENV", "testing")

	log.Logger = logger.Make(zerolog.TraceLevel)
	t.Log("test")

	var m = map[string]any{
		"bot_id": 1,
		"result": "ok",
	}

	b, err := json.Marshal(&m)
	assert.NoError(t, err)
	s := &Subscriber{}
	s.do(bots, b)

}
