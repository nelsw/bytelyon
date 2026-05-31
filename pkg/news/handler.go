package news

import (
	"net/http"
	"slices"

	"github.com/nelsw/bytelyon/pkg/api"
	"github.com/nelsw/bytelyon/pkg/article"
	"github.com/nelsw/bytelyon/pkg/page"
)

func Handler(r api.HTTPRequest) api.HTTPResponse {
	switch r.RequestContext.HTTP.Method {
	case http.MethodGet:
		return HandleGet(r)
	case http.MethodDelete:
		return HandleDelete(r)
	}
	return api.NotImplemented()
}

func HandleDelete(r api.HTTPRequest) api.HTTPResponse {

	arr := Find(r.UserID(), r.Query("topic"))
	idx := -1
	for i, a := range arr {
		if a.URL == r.Query("url") {
			idx = i
			break
		}
	}

	if idx >= 0 {
		headline := arr[idx]
		if err := page.Delete(headline.URL, headline.ID); err != nil {
			return api.BadRequest(err)
		}

		if err := Save(r.UserID(), r.Query("topic"), slices.Delete(arr, idx, idx+1)); err != nil {
			return api.BadRequest(err)
		}
	}

	return api.NoContent()
}

func HandleGet(r api.HTTPRequest) api.HTTPResponse {
	arr := Find(r.UserID(), r.Query("topic"))
	if r.Query("url") == "" {
		return api.OK(arr)
	}

	for _, h := range arr {
		if h.URL != r.Query("url") {
			continue
		}
		if a, err := article.Find(h.URL, h.ID); err == nil {
			return api.OK(a)
		}
		break
	}
	return api.NoContent()
}
