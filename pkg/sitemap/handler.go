package sitemap

import (
	"net/http"

	"github.com/nelsw/bytelyon/pkg/api"
	"github.com/nelsw/bytelyon/pkg/snippet"
)

func Handler(r api.Request) api.Response {
	switch r.Method() {
	case http.MethodGet:
		return HandleGet(r)
	}
	return r.NI()
}

func HandleGet(r api.Request) api.Response {
	if r.Query("url") == "" {
		return r.OK(Find(r.UserID(), r.Query("domain")))
	}
	return r.OK(snippet.Find(r.Query("url")))
}
