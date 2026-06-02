package profile

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/nelsw/bytelyon/pkg/api"
)

func Handler(r api.HTTPRequest) api.HTTPResponse {
	switch r.RequestContext.HTTP.Method {
	case http.MethodGet:
		return handleGet(r)
	case http.MethodPut:
		return handlePut(r)
	}
	return api.NotImplemented()
}

// HandleGet returns a user profile with redacted credentials
func handleGet(r api.HTTPRequest) api.HTTPResponse {
	m, err := Find(r.UserID())
	if err != nil {
		return api.BadRequest(err)
	}
	return api.OK(map[string]any{
		"anthropic": m.Anthropic.ApiKey != "",
		"img":       m.Img,
		"shopify": map[string]any{
			"store": m.Shopify.Store,
			"client": map[string]any{
				"id":     m.Shopify.Client.ID != "",
				"secret": m.Shopify.Client.Secret != "",
			},
		},
		"verified": m.Verified,
	})
}

// handlePut saves a user profile
func handlePut(r api.HTTPRequest) api.HTTPResponse {
	var m Model
	if err := json.Unmarshal([]byte(r.Body), &m); err != nil {
		return api.BadRequest(err)
	} else if err = Save(r.UserID(), &m); err != nil {
		return api.BadRequest(err)
	}
	fmt.Printf("Saved model %+v\n", m)
	return api.NoContent()
}
