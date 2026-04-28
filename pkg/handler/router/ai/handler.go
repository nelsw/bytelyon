package ai

import (
	"encoding/json"
	"net/http"

	"github.com/nelsw/bytelyon/pkg/api"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/service"
	"github.com/rs/zerolog/log"
)

func Handler(r api.Request) api.Response {
	switch r.Method() {
	case http.MethodPost:
		return handlePostArticle(r)
	}
	return r.NI()
}

func handlePostArticle(r api.Request) api.Response {
	var a = new(model.Article)
	if err := json.Unmarshal([]byte(r.Body), a); err != nil {
		log.Err(err).Msg("failed to unmarshal article")
		return r.BAD(err)
	}

	link, err := service.SpinArticle(a)
	if err != nil {
		return r.BAD(err)
	}
	return r.OK(map[string]any{"link": link})
}
