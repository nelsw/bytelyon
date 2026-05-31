package ai

import (
	"encoding/json"
	"net/http"

	"github.com/nelsw/bytelyon/pkg/api"
	"github.com/rs/zerolog/log"
)

type prompt struct {
	System  string `json:"system"`
	Message string `json:"message"`
	HTML    bool   `json:"html"`
}

func Handler(r api.HTTPRequest) api.HTTPResponse {
	if r.IsGuest() {
		return api.Forbidden()
	}
	switch r.RequestContext.HTTP.Method {
	case http.MethodPost:
		return handlePost(r)
	}
	return api.NotImplemented()
}

func handlePost(r api.HTTPRequest) api.HTTPResponse {
	var p prompt
	if err := json.Unmarshal([]byte(r.Body), &p); err != nil {
		log.Err(err).Msg("failed to unmarshal prompt")
		return api.BadRequest(err)
	}

	txt, err := Prompt(p.System, p.Message, p.HTML)
	if err != nil {
		return api.BadRequest(err)
	}
	return api.OK(map[string]string{"text": txt})
}
