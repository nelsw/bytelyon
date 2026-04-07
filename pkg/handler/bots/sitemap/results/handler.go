package bots

import (
	"net/http"
	"sort"

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
	results := repo.FindBotResults(r.UserID(), r.ID(), model.SitemapBotType)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Timestamp().Compare(results[j].Timestamp()) == -1
	})
	return r.OK(results)
}
