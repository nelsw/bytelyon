package sitemap

import (
	"net/http"

	"github.com/nelsw/bytelyon/pkg/api"
	"github.com/nelsw/bytelyon/pkg/page"
	"github.com/nelsw/bytelyon/pkg/snippet"
)

func Handler(r api.Request) api.Response {
	switch r.Method() {
	case http.MethodGet:
		return HandleGet(r)
	case http.MethodDelete:
		return HandleDelete(r)
	}
	return r.NI()
}

func HandleDelete(r api.Request) api.Response {
	if err := Delete(r.UserID(), r.Domain()); err != nil {
		return r.BAD(err)
	}
	return r.NC()
}

func HandleGet(r api.Request) api.Response {
	if r.URL() == "" {
		arr, err := Find(r.UserID(), r.Domain())
		if err != nil {
			return r.BAD(err)
		}
		return r.OK(arr)
	}

	out, err := page.FindObjects[snippet.Model](r.URL())
	if err != nil {
		return r.BAD(err)
	}
	return r.OK(out)
}
