package news

import (
	"net/http"

	"github.com/nelsw/bytelyon/pkg/api"
	"github.com/nelsw/bytelyon/pkg/repo"
)

func Handler(r api.Request) api.Response {
	switch r.Method() {
	case http.MethodGet:
		return handleGet(r)
	case http.MethodDelete:
		return HandleDelete(r)
	}
	return r.NI()
}

func handleGet(r api.Request) api.Response {
	out, err := repo.GetNews(r.UserID(), r.BotID())
	if err != nil {
		return r.BAD(err)
	}
	return r.OK(out)
}

func HandleDelete(r api.Request) api.Response {
	if err := repo.DeleteNews(r.UserID(), r.BotID(), r.Query("url")); err != nil {
		return r.BAD(err)
	}
	return r.NC()
}
