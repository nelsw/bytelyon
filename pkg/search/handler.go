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

	if r.ID().IsZero() {
		if err := Delete(r.UserID(), r.Query("query")); err != nil {
			return r.BAD(err)
		}
		return r.NC()
	}

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

	m, err := Find(r.UserID(), r.Query("query"))
	if err != nil {
		return r.BAD(err)
	}

	if r.ID().IsZero() {
		return r.OK(m)
	}

	val, ok := m[r.ID()]
	if !ok {
		return r.NC()
	}
	return r.OK(val)
}
