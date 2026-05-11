package em

import (
	"time"

	"github.com/nelsw/bytelyon/pkg/entity"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/store"
	"github.com/oklog/ulid/v2"
)

func searchDB(userID ulid.ULID) (*store.DB[string, *entity.Search], error) {
	return store.New[string, *entity.Search]("users", userID, "bots", model.SearchBotType)
}

func GetSearch(userID ulid.ULID, domain string) (*entity.Search, bool) {
	db, err := searchDB(userID)
	if err != nil {
		return nil, false
	}
	defer db.Close()
	return db.Get(domain)
}

func SaveSearch(userID ulid.ULID, e *entity.Search) {

	db, err := searchDB(userID)
	if err != nil {
		return
	}
	defer db.Close()

	db.Set(e.ID.Timestamp().Format(time.RFC3339), e)
	for _, p := range e.Pages.Values() {
		SavePage(p)
	}
}

func DeleteSearch(userID ulid.ULID, domain string) {
	db, err := searchDB(userID)
	if err != nil {
		return
	}
	defer db.Close()
	db.Drop(domain)
}
