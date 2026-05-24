package searches

import (
	"net/http"

	"github.com/nelsw/bytelyon/pkg/api"
	"github.com/nelsw/bytelyon/pkg/entity"
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
	if e := new(entity.Search).Find(r.UserID(), r.Query("query")); e != nil {
		return r.OK(e)
	}
	return r.NC()
}

func handleDelete(r api.Request) api.Response {
	new(entity.Search).Delete(r.UserID(), r.Query("query"))
	return r.NC()
}
