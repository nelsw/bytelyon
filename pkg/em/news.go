package em

import (
	"github.com/nelsw/bytelyon/pkg/entity"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/store"
	"github.com/oklog/ulid/v2"
)

func newsDB(userID ulid.ULID) (*store.DB[string, *entity.News], error) {
	return store.New[string, *entity.News]("users", userID, "bots", model.NewsBotType)
}

func SaveNews(userID ulid.ULID, e *entity.News) {
	db, err := newsDB(userID)
	if err != nil {
		return
	}
	defer db.Close()
	db.Set(topic, news)
}

func GetNews(userID ulid.ULID, topic string) (*entity.News, bool) {
	db, err := newsDB(userID)
	if err != nil {
		return nil, false
	}
	defer db.Close()
	return db.Get(topic)
}

func DeleteNews(userID ulid.ULID, topic string) {
	db, err := newsDB(userID)
	if err != nil {
		return
	}
	defer db.Close()
	db.Drop(topic)
}
