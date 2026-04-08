package router

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"

	. "github.com/nelsw/bytelyon/pkg/api"
	"github.com/nelsw/bytelyon/pkg/db"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/repo"
	"github.com/nelsw/bytelyon/pkg/service"
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
	case http.MethodPost:
		return handlePost(r), nil
	}

	return r.NI(), nil
}

// handleDelete deletes a bot or bot result from the database using the following routes:
//
//	bot: /bots?type=...&target=...
//
// result: /bots?type=...&target=...&id=...
func handleDelete(r Request) Response {

	var err error

	if r.ID().IsZero() {
		err = repo.DeleteBot(r.UserID(), r.Target(), r.BotType())
	} else {
		err = repo.DeleteBotResult(r.UserID(), r.BotID(), r.ID(), r.BotType())
	}

	if err != nil {
		return r.BAD(err)
	}
	return r.NC()
}

// handleGet queries the database for bots and bot results using the following routes:
// - bots: /bots?type=...
// - results: /bots?type=...&id=...
func handleGet(r Request) Response {

	// if the request is for bots
	if r.ID().IsZero() {
		bots := repo.FindBotsByType(r.UserID(), r.BotType())
		sort.Slice(bots, func(i, j int) bool {
			return strings.Compare(bots[i].Target, bots[j].Target) == -1
		})
		return r.OK(bots)
	}

	// else the request is for bot results
	results := repo.FindBotResults(r.UserID(), r.ID(), r.BotType())
	if r.BotType() == model.SitemapBotType {
		return r.OK(model.NewSitemapResults(results))
	}
	if r.BotType() == model.SearchBotType {
		return r.OK(results.ToNodes())
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Timestamp().Compare(results[j].Timestamp()) == -1
	})
	return r.OK(results)
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

func handlePost(r Request) Response {

	if !r.IsStu() {
		return r.NOPE()
	}

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
