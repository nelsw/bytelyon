package model

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResult_UnmarshalJSON(t *testing.T) {
	var m = map[string]any{
		"bot_id": 1,
		"result": "ok",
	}

	b, err := json.Marshal(&m)
	assert.NoError(t, err)

	var r Result
	err = json.Unmarshal(b, &r)
	assert.NoError(t, err)

	assert.Equal(t, m["bot_id"], r.BotID)
	assert.Equal(t, m["result"], r.Result)
}
