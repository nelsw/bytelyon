package model

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nelsw/bytelyon/apps/mux/internal/util/uuid"
	"github.com/rs/zerolog/log"
)

type Search struct {
	BotID         int    `json:"bot_id"`
	ContentKey    string `json:"content_key"`
	Data          any    `json:"data"`
	Query         string `json:"query"`
	ScreenshotKey string `json:"screenshot_key"`
	screenshot    []byte
	content       []byte
}

func (s *Search) Screenshot() (string, []byte, bool) {
	return s.ScreenshotKey, s.screenshot, true
}

func (s *Search) Content() (string, []byte, bool) {
	return s.ContentKey, s.content, true
}

func (s *Search) Route() string {
	return fmt.Sprintf("bots/%d/searches", s.BotID)
}

func (s *Search) UnmarshalJSON(data []byte) (err error) {
	defer func() {
		if err != nil {
			log.Err(err).Bytes("data", data).Msg("Search:UnmarshalJSON")
		}
	}()

	var v struct {
		ID             int      `json:"id"`
		BotID          int      `json:"bot_id"`
		SimilarQueries []string `json:"similar_queries"`
		Query          string   `json:"query"`
		Screenshot     []byte   `json:"screenshot"`
		Content        []byte   `json:"content"`
	}

	if err = json.Unmarshal(data, &v); err != nil {
		return
	}

	url := "https://www.google.com?q=" + strings.ReplaceAll(v.Query, " ", "+")
	uid := uuid.FromURL(url)

	*s = Search{
		BotID:         v.BotID,
		ContentKey:    fmt.Sprintf("%s/%d/%s/content.html", SearchBot, v.ID, uid),
		Data:          map[string]any{"similar_queries": v.SimilarQueries},
		Query:         v.Query,
		ScreenshotKey: fmt.Sprintf("%s/%d/%s/screenshot.png", SearchBot, v.ID, uid),
	}

	return
}
