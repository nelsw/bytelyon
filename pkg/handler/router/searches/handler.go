package searches

import (
	"net/http"

	"github.com/nelsw/bytelyon/pkg/api"
	"github.com/nelsw/bytelyon/pkg/em"
)

func Handler(r api.Request) api.Response {
	switch r.Method() {
	case http.MethodGet:
		return handleGet(r)
	case http.MethodDelete:
		return handleDelete(r)
	}
	return r.NI()
}

func handleGet(r api.Request) api.Response {
	if a, ok := em.GetSitemap(r.UserID(), r.Target()); ok {
		return r.OK(a)
	}
	return r.NC()
}

func handleDelete(r api.Request) api.Response {
	em.DeleteSitemap(r.UserID(), r.Target())
	return r.NC()
}
