package search

import (
	"net/http"

	"github.com/nelsw/bytelyon/pkg/api"
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

	m, err := Find(r.UserID(), r.Query("query"))
	if err != nil {
		return r.BAD(err)
	}

	delete(m, r.ID())

	if err = Save(r.UserID(), r.Query("query"), m); err != nil {
		return r.BAD(err)
	}
	return r.NC()
}

func HandleGet(r api.Request) api.Response {

	if r.ID().IsZero() {
		arr, err := FindIDs(r.UserID(), r.Query("query"))
		if err != nil {
			return r.BAD(err)
		}
		return r.OK(arr)
	}

	arr, err := FindSerp(r.UserID(), r.Query("query"), r.ID())
	if err != nil {
		return r.BAD(err)
	}
	return r.OK(arr)
}
