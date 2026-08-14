package model

import (
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog/log"
)

type Sitemap struct {
	BotID  int      `json:"bot_id"`
	Domain string   `json:"domain"`
	URLS   []string `json:"urls"`
}

func (s *Sitemap) Route() string {
	return fmt.Sprintf("bots/%d/sitemaps", s.BotID)
}

func (s *Sitemap) UnmarshalJSON(data []byte) (err error) {
	defer func() {
		if err != nil {
			log.Err(err).Bytes("data", data).Msg("Sitemap:UnmarshalJSON")
		}
	}()

	var v struct {
		BotID  int      `json:"bot_id"`
		Domain string   `json:"domain"`
		URLS   []string `json:"urls"`
	}

	if err = json.Unmarshal(data, &v); err == nil {
		*s = v
	}

	return
}
