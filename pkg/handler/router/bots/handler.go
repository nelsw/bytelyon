package bots

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"

	. "github.com/nelsw/bytelyon/pkg/api"
	"github.com/nelsw/bytelyon/pkg/db"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/repo"
	"github.com/nelsw/bytelyon/pkg/urls"
	"github.com/rs/zerolog/log"
)

func Handler(r Request) Response {

	r.Log()

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
	if err := repo.DeleteBot(r.UserID(), r.Target(), r.BotType()); err != nil {
		return r.BAD(err)
	}
	return r.NC()
}

// handleGet queries the database for bots.
func handleGet(r Request) Response {
	bots := repo.FindBotsByType(r.UserID(), r.BotType())
	sort.Slice(bots, func(i, j int) bool {
		return strings.Compare(bots[i].Target, bots[j].Target) == -1
	})
	return r.OK(bots)
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
	b.Target = strings.ToLower(b.Target)

	if b.Type == model.SitemapBotType {
		b.Target = urls.Domain(b.Target)
	}

	if err := db.Put(b); err != nil {
		log.Err(err).Msg("failed to put bot")
		return r.BAD(err)
	}
	log.Debug().Object("bot", b).Msg("bot put")

	return r.OK(b)
}
