package sitemaps

import (
	"net/http"

	"github.com/nelsw/bytelyon/pkg/api"
	"github.com/nelsw/bytelyon/pkg/em"
	"github.com/nelsw/bytelyon/pkg/entity"
)

func Handler(r api.Request) api.Response {

	e := entity.NewSitemap(r.UserID(), r.Query("domain"))
	if err := em.Find(e); err != nil {
		return r.BAD(err)
	}

	switch r.Method() {
	case http.MethodGet:
		return r.OK(e)
	case http.MethodDelete:
		return r.OF(em.Delete(e))
	}

	return r.NI()
}
