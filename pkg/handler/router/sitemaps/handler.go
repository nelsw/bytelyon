package sitemaps

import (
	"net/http"

	"github.com/nelsw/bytelyon/pkg/api"
	"github.com/nelsw/bytelyon/pkg/model"
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
	if e := new(model.Sitemap).Find(r.UserID(), r.Query("domain")); e != nil {
		return r.OK(e)
	}
	return r.NC()
}

func handleDelete(r api.Request) api.Response {
	new(model.Sitemap).Delete(r.UserID(), r.Query("domain"))
	return r.NC()
}
