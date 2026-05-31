package search

import (
	"net/http"
	"slices"

	"github.com/nelsw/bytelyon/pkg/api"
	"github.com/nelsw/bytelyon/pkg/id"
)

func Handler(r api.HTTPRequest) api.HTTPResponse {
	switch r.RequestContext.HTTP.Method {
	case http.MethodGet:
		return HandleGet(r)
	case http.MethodDelete:
		return HandleDelete(r)
	}
	return api.NotImplemented()
}

func HandleDelete(r api.HTTPRequest) api.HTTPResponse {

	arr, err := Find(r.UserID(), r.Query("query"))
	if err != nil {
		return api.BadRequest(err)
	}
	qid := id.ParseULID(r.Query("id"))
	idx := -1
	for i, a := range arr {
		if a.Compare(qid) == 0 {
			idx = i
			break
		}
	}

	if idx < 0 {
		return api.NoContent()
	}

	arr = slices.Delete(arr, idx, idx+1)

	if err = Save(r.UserID(), r.Query("query"), arr); err != nil {
		return api.BadRequest(err)
	}
	return api.NoContent()
}

func HandleGet(r api.HTTPRequest) api.HTTPResponse {
	qid := id.ParseULID(r.Query("id"))
	if qid.IsZero() {
		arr, err := Find(r.UserID(), r.Query("query"))
		if err != nil {
			return api.BadRequest(err)
		}
		return api.OK(arr)
	}

	arr, err := FindSerp(r.UserID(), r.Query("query"), qid)
	if err != nil {
		return api.BadRequest(err)
	}
	return api.OK(arr)
}
