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

func Handler(r api.Request) api.Response {
	if r.IsGuest() {
		return r.NOPE()
	}
	switch r.RequestContext.HTTP.Method {
	case http.MethodPost:
		return handlePost(r)
	}
	return r.NI()
}

func handlePost(r api.Request) api.Response {
	var p prompt
	if err := json.Unmarshal([]byte(r.Body), &p); err != nil {
		log.Err(err).Msg("failed to unmarshal prompt")
		return r.BAD(err)
	}

	txt, err := Prompt(p.System, p.Message, p.HTML)
	if err != nil {
		return r.BAD(err)
	}
	return r.OK(map[string]string{"text": txt})
}
