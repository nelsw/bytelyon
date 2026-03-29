package router

import (
	"encoding/json"
	"net/http"

	. "github.com/nelsw/bytelyon/pkg/api"
	"github.com/nelsw/bytelyon/pkg/db"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/repo"
	"github.com/rs/zerolog/log"
)

func Handler(r Request) (Response, error) {

	r.Log()

	switch r.Method() {
	case http.MethodDelete:
		return handleDelete(r), nil
	case http.MethodGet:
		return handleGet(r), nil
	case http.MethodPut:
		return handlePut(r), nil
	}

	return r.NI(), nil
}

// handleDelete deletes a bot or bot result from the database using the following routes:
//
//	bot: /bots?type=...&target=...
//
// result: /bots?type=...&target=...&id=...
func handleDelete(r Request) Response {

	if r.ID().IsZero() {
		log.Info().Msg("deleting bot and results")
		if err := repo.DeleteBot(r.UserID(), r.Target(), r.BotType()); err != nil {
			log.Error().Err(err).Msg("failed to delete bot")
			return r.BAD(err)
		}
		return r.NC()
	}

	log.Info().Msg("deleting bot result")
	if err := repo.DeleteBotResult(r.UserID(), r.ID(), r.BotType()); err != nil {
		log.Error().Err(err).Msg("failed to delete bot result")
		return r.BAD(err)
	}
	return r.NC()
}

// handleGet queries the database for bots and bot results using the following routes:
// - bots: /bots?type=...
// - results: /bots?type=...&id=...
func handleGet(r Request) Response {

	var nodes model.Nodes

	if r.ID().IsZero() {
		nodes = repo.
			FindBotsByType(r.UserID(), r.BotType()).
			ToNodes()
	} else {
		nodes = repo.
			FindBotResults(r.UserID(), r.ID(), r.BotType()).
			ToNodes()
	}

	if len(nodes) == 0 {
		return r.NC()
	}

	return r.OK(nodes)
}

// handlePut creates or updates a bot in the database for the given body.
func handlePut(r Request) Response {

	var b = new(model.Bot)
	if err := json.Unmarshal([]byte(r.Body), b); err != nil {
		log.Err(err).Msg("failed to unmarshal bot")
		return r.BAD(err)
	}
	log.Debug().Object("bot", b).Msg("bot unmarshalled")

	if err := b.Validate(); err != nil {
		log.Err(err).Msg("failed to validate bot")
		return r.BAD(err)
	}
	log.Debug().Object("bot", b).Msg("bot validated")

	b.UserID = r.UserID()
	if b.ID.IsZero() {
		b.ID = model.NewULID()
	}

	if err := db.PutItem(b); err != nil {
		log.Err(err).Msg("failed to put bot")
		return r.BAD(err)
	}
	log.Debug().Object("bot", b).Msg("bot put")

	return r.OK(b)
}
