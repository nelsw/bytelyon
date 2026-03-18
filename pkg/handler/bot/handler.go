package bot

import (
	"encoding/json"
	"net/http"

	. "github.com/nelsw/bytelyon/pkg/api"
	"github.com/nelsw/bytelyon/pkg/db"
	"github.com/nelsw/bytelyon/pkg/model"
)

// api/bots/type/:type
// api/bots/type/:type/target/:target
// api/bots/type/:type/target/:target/id/:id

// Handler
func Handler(r Request) (Response, error) {

	r.Log()

	// todo - validation

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

func handleDelete(r Request) Response {
	err := db.Delete(&model.Bot{
		UserID: r.UserID(),
		Target: r.Query("target"),
		Type:   model.BotType(r.Query("type")),
	})
	if err != nil {
		return r.BAD(err)
	}
	return r.NC()
}

func handleGet(r Request) Response {

	bots, err := db.Query(&model.Bot{
		UserID: r.UserID(),
		Type:   model.BotType(r.Query("type")),
	})
	if err != nil {
		return r.BAD(err)
	}

	if r.Query("id") == "" {
		return r.OK(bots)
	}

	var results []*model.BotResult
	for _, bot := range bots {
		if bot.ID.String() != r.Query("id") {
			continue
		}
		results, err = db.Query(&model.BotResult{
			BotID: bot.ID,
			Type:  bot.Type,
		})
		break
	}

	if err != nil {
		return r.BAD(err)
	}

	return r.OK(results)
}

func handlePut(r Request) Response {

	var b model.Bot
	if err := json.Unmarshal([]byte(r.Body), &b); err != nil {
		return r.BAD(err)
	} else if err = b.Validate(); err != nil {
		return r.BAD(err)
	}
	b.UserID = r.UserID()
	if b.ID.IsZero() {
		b.ID = model.NewULID()
	}

	if err := db.PutItem(&b); err != nil {
		return r.BAD(err)
	}

	return r.OK(&b)
}
