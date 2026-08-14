package job

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/nelsw/bytelyon/apps/mux/internal/model"
	"github.com/rs/zerolog/log"
)

func WebHandler(q *Queue) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var bot model.Bot
		if err := json.NewDecoder(r.Body).Decode(&bot); err != nil {
			log.Warn().Err(err).Msg("Failed to decode bot")
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			q.Send(&bot)
			w.WriteHeader(http.StatusCreated)
		}
	}
}

func BroHandler(q *Queue) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if botID, err := strconv.Atoi(r.PathValue("id")); err != nil {
			log.Warn().Err(err).Msgf("Invalid bot ID: %s", r.PathValue("id"))
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			q.Del(botID)
			w.WriteHeader(http.StatusNoContent)
		}
	}
}