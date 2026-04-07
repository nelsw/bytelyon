package news

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
	}

	return r.NI(), nil
}

// handleGet queries the database for bots and bot results using the following routes:
// - bots: /bots/news
// - results: /bots?type=...&id=...
func handleGet(r Request) Response {

	if r.ID().IsZero() {
		bots := repo.FindBotsByType(r.UserID(), model.NewsBotType)
		sort.Slice(bots, func(i, j int) bool {
			return strings.Compare(bots[i].Target, bots[j].Target) == -1
		})
		return r.OK(bots)
	}

	results := repo.FindBotResults(r.UserID(), r.ID(), model.NewsBotType)

	return r.OK(results)
}
