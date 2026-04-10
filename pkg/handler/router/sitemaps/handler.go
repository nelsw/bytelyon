package sitemaps

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
	m, err := db.Get(&model.Sitemap{Domain: r.Domain()})
	if err != nil {
		return r.BAD(err)
	}
	return r.OK(m)
}
