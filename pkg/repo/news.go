package repo

import (
	"github.com/nelsw/bytelyon/pkg/store"
	"github.com/oklog/ulid/v2"
)

func GetNews(userID, botID ulid.ULID) ([]any, error) {
	db, err := store.New[string, any](userID, botID)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	return db.Values(), nil
}

func DeleteNews(userID, botID ulid.ULID, url string) error {
	db, err := store.New[string, any](userID, botID)
	if err != nil {
		return err
	}
	defer db.Close()
	db.Drop(url)
	return nil
}
