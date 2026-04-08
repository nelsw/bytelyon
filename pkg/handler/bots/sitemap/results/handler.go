package bots

import (
	"net/http"

	. "github.com/nelsw/bytelyon/pkg/api"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/repo"
)

func Handler(r Request) (Response, error) {

	r.Log()

	switch r.Method() {
	case http.MethodGet:
		return handleGet(r), nil
	}

	return r.NI(), nil
}

func handleGet(r Request) Response {

	// find results, fail fast if empty
	results := repo.FindBotResults(r.UserID(), r.ID(), model.SitemapBotType)
	if len(results) == 0 {
		return r.NC()
	}

	return r.OK(model.NewSitemapResults(results))
}
