package bot

import (
	"encoding/json"
	"net/http"

	"github.com/nelsw/bytelyon/pkg/api"
)

func Handler(r api.HTTPRequest) api.HTTPResponse {
	switch r.RequestContext.HTTP.Method {
	case http.MethodDelete:
		return handleDelete(r)
	case http.MethodGet:
		return handleGet(r)
	case http.MethodPost:
		return HandlePost(r)
	case http.MethodPut:
		return handlePut(r)
	}
	return api.NotImplemented()
}

func handleDelete(r api.HTTPRequest) api.HTTPResponse {
	if err := Delete(r.UserID(), Type(r.Query("type")), r.Query("target")); err != nil {
		return api.BadRequest(err)
	}
	return api.NoContent()
}

func handleGet(r api.HTTPRequest) api.HTTPResponse {
	if all := FindAll(r.UserID(), Type(r.Query("type"))); len(all) > 0 {
		return api.OK(all)
	}
	return api.NoContent()
}

func HandlePost(r api.HTTPRequest) api.HTTPResponse {
	var m = new(Model)
	if err := json.Unmarshal([]byte(r.Body), m); err != nil {
		return api.BadRequest(err)
	} else if err = Create(r.UserID(), m); err != nil {
		return api.BadRequest(err)
	}
	return api.OK(m)
}

func handlePut(r api.HTTPRequest) api.HTTPResponse {
	var m = new(Model)
	if err := json.Unmarshal([]byte(r.Body), m); err != nil {
		return api.BadRequest(err)
	} else if err = Update(r.UserID(), m); err != nil {
		return api.BadRequest(err)
	}
	return api.OK(m)
}
