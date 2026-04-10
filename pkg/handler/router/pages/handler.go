package pages

import (
	"net/http"

	"github.com/nelsw/bytelyon/pkg/api"
	"github.com/nelsw/bytelyon/pkg/db"
	"github.com/nelsw/bytelyon/pkg/model"
)

func Handler(r api.Request) api.Response {
	switch r.Method() {
	case http.MethodGet:
		return handleGet(r)
	}
	return r.NI()
}

func handleGet(r api.Request) api.Response {
	m, err := db.Query(&model.Page{URL: r.Query("url")})
	if err != nil {
		return r.BAD(err)
	}
	return r.OK(m)
}
