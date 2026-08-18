package model

import (
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog/log"
)

type Result struct {
	BotID  int    `json:"bot_id"`
	Result string `json:"result"`
}

func (r *Result) Route() string {
	return fmt.Sprintf("/bots/%d", r.BotID)
}

func (r *Result) UnmarshalJSON(data []byte) (err error) {
	defer func() {
		if err != nil {
			log.Err(err).Bytes("data", data).Msg("Result:UnmarshalJSON")
		}
	}()

	var v struct {
		BotID  int    `json:"bot_id"`
		Result string `json:"result"`
	}

	if err = json.Unmarshal(data, &v); err == nil {
		*r = v
	}

	return
}
