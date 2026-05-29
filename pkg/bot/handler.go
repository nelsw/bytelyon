package bot

import (
	"encoding/json"
	"net/http"

	. "github.com/nelsw/bytelyon/pkg/api"
	"github.com/rs/zerolog/log"
)

func Handler(r Request) Response {
	switch r.Method() {
	case http.MethodDelete:
		return handleDelete(r)
	case http.MethodGet:
		return handleGet(r)
	case http.MethodPut:
		return handlePut(r)
	}
	return r.NI()
}

func handleDelete(r Request) Response {
	if err := Delete(r.UserID(), r.Query("type"), r.Query("target")); err != nil {
		return r.BAD(err)
	}
	return r.NC()
}

func handleGet(r Request) Response { return r.OK(Find(r.UserID(), r.Query("type"))) }

func handlePut(r Request) Response {

	var m = new(Model)
	if err := json.Unmarshal([]byte(r.Body), m); err != nil {
		log.Err(err).Msg("failed to unmarshal bot")
		return r.BAD(err)
	}

	if err := Save(r.UserID(), m); err != nil {
		log.Err(err).Msg("failed to save bot")
		return r.BAD(err)
	}

	return r.OK(m)
}
