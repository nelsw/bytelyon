package news

import (
	"net/http"

	"github.com/nelsw/bytelyon/pkg/api"
)

func Handler(r api.Request) api.Response {

	µ := New(r.UserID(), r.Query("topic"))
	if !µ.Find(true) {
		return r.NC()
	}

	switch r.Method() {
	case http.MethodGet:
		return r.OK(µ)
	case http.MethodDelete:
		return r.OF(µ.Delete())
	}

	return r.NI()
}
