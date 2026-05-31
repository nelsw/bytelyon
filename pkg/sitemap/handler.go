package sitemap

import (
	"net/http"

	"github.com/nelsw/bytelyon/pkg/api"
	"github.com/nelsw/bytelyon/pkg/snippet"
)

func Handler(r api.HTTPRequest) api.HTTPResponse {
	switch r.RequestContext.HTTP.Method {
	case http.MethodGet:
		return HandleGet(r)
	}
	return api.NoContent()
}

func HandleGet(r api.HTTPRequest) api.HTTPResponse {
	if r.Query("url") == "" {
		return api.OK(Find(r.UserID(), r.Query("domain")))
	}
	return api.OK(snippet.Find(r.Query("url")))
}
