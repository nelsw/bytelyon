package search

import (
	"net/http"
	"slices"

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

	arr, err := Find(r.UserID(), r.Query("query"))
	if err != nil {
		return r.BAD(err)
	}

	idx := -1
	for i, a := range arr {
		if a.Compare(r.ID()) == 0 {
			idx = i
			break
		}
	}

	if idx < 0 {
		return r.NC()
	}

	slices.Delete(arr, idx, idx+1)

	if err = Save(r.UserID(), r.Query("query"), arr); err != nil {
		return r.BAD(err)
	}
	return r.NC()
}

func HandleGet(r api.Request) api.Response {

	if r.ID().IsZero() {
		arr, err := Find(r.UserID(), r.Query("query"))
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
