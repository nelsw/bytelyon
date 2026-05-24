package pages

import (
	"net/http"

	"github.com/nelsw/bytelyon/pkg/api"
	"github.com/nelsw/bytelyon/pkg/entity"
	"github.com/oklog/ulid/v2"
)

func Handler(r api.Request) api.Response {
	switch r.Method() {
	case http.MethodGet:
		return handleGet(r)
	}
	return r.NI()
}

func handleGet(r api.Request) api.Response {

	id, err := ulid.Parse(r.Query("id"))
	if err != nil {
		return r.BAD(err)
	}

	if e := entity.FindPage(r.Query("url"), id); e != nil {
		return r.OK(e)
	}

	return r.NC()
}
