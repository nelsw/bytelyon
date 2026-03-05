package bot

import (
	"encoding/base64"
	"net/http"

	. "github.com/nelsw/bytelyon/pkg/api"
	"github.com/nelsw/bytelyon/pkg/db"
	. "github.com/nelsw/bytelyon/pkg/model"
)

func Handler(r Request) (Response, error) {

	r.Log()

	switch r.Method() {
	case http.MethodDelete:
		return handleDelete(r), nil
	case http.MethodGet:
		return handleGet(r), nil
	case http.MethodPost:
		fallthrough
	case http.MethodPut:
		return handlePut(r), nil
	}

	return r.NI(), nil
}

func handleDelete(r Request) Response {

	bot := Bot{
		UserID: r.UserID(),
		Target: r.Param("target"),
		Type:   BotType(r.Param("type")),
	}

	if err := bot.Validate(); err != nil {
		return r.BAD(err)
	}

	if bot.Type.IsSitemap() {
		tgt, err := r.Base64Param("target", base64.URLEncoding)
		if err != nil {
			return r.BAD(err)
		}
		bot.Target = string(tgt)
	}

	if err := db.Delete(bot); err != nil {
		return r.BAD(err)
	}
	return r.NC()
}

func handleGetBot(r Request) Response {

	bot := Bot{
		UserID: r.UserID(),
		Type:   BotType(r.Param("type")),
	}

	if err := bot.Validate(); err != nil {
		return r.BAD(err)
	}

	arr, err := db.Query[Bot](bot)
	if err != nil {
		return r.BAD(err)
	}
	return r.OK(arr)
}

func handlePut(r Request) Response {

	in := Bot{
		UserID: r.UserID(),
		Target: r.Param("target"),
		Type:   BotType(r.Param("type")),
	}

	if err := in.Validate(); err != nil {
		return r.BAD(err)
	}

	out, err := Body[Bot](r, in)
	if err != nil {
		return r.BAD(err)
	}
	out.UserID = in.UserID
	out.Target = in.Target
	out.Type = in.Type

	if err = db.Put(out); err != nil {
		return r.BAD(err)
	}
	return r.OK(out)
}
