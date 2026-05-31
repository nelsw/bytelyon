package bot

import (
	"encoding/json"
	"net/http"

	"github.com/nelsw/bytelyon/pkg/api"
	"github.com/rs/zerolog/log"
)

func Handler(r api.Request) api.Response {
	switch r.RequestContext.HTTP.Method {
	case http.MethodDelete:
		return handleDelete(r)
	case http.MethodGet:
		return handleGet(r)
	case http.MethodPut:
		return handlePut(r)
	}
	return r.NI()
}

func handleDelete(r api.Request) api.Response {
	if err := Delete(r.UserID(), Type(r.Query("type")), r.Query("target")); err != nil {
		return r.BAD(err)
	}
	return r.NC()
}

func handleGet(r api.Request) api.Response { return r.OK(Find(r.UserID(), Type(r.Query("type")))) }

func handlePut(r api.Request) api.Response {

	var m = new(Model)
	if err := json.Unmarshal([]byte(r.Body), m); err != nil {
		return r.BAD(err)
	}

	if err := Save(r.UserID(), m); err != nil {
		log.Err(err).Msg("failed to save bot")
		return r.BAD(err)
	}

	return r.OK(m)
}
