package bots

import (
	"net/http"
	"sort"
	"strings"

	. "github.com/nelsw/bytelyon/pkg/api"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/repo"
)

func Handler(r Request) (Response, error) {

	r.Log()

	switch r.Method() {
	case http.MethodGet:
		return handleGet(r), nil
	case http.MethodPost:
		return handlePost(r), nil
	case http.MethodPut:
		return handlePut(r), nil
	case http.MethodDelete:
		return handleDelete(r), nil
	}

	return r.NI(), nil
}

func handleGet(r Request) Response {
	bots := repo.FindBotsByType(r.UserID(), model.NewsBotType)
	sort.Slice(bots, func(i, j int) bool {
		return strings.Compare(bots[i].Target, bots[j].Target) == -1
	})
	return r.OK(bots)
}

func handlePost(r Request) Response {
	return r.NC()
}

func handlePut(r Request) Response {
	return r.NC()
}

func handleDelete(r Request) Response {
	return r.NC()
}
