package entity

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"

	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/s3"
	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/oklog/ulid/v2"
)

type News struct {
	Topic    string
	Articles map[string]model.Article
	userID   ulid.ULID
}

func (e *News) key() string {
	return fmt.Sprintf("users/%s/news/%s.json", e.userID, e.Topic)
}

func (e *News) Save() {
	s3.PutPrivateObject(e.key(), util.JSON(e))
}

func (e *News) From(userID ulid.ULID, topic string) *News {
	if x := e.Find(userID, topic); x != nil {
		return x
	}
	return e.Create(userID, topic)
}

func (e *News) Create(userID ulid.ULID, topic string) *News {
	x := &News{
		Articles: make(map[string]model.Article),
		Topic:    topic,
		userID:   userID,
	}
	x.Save()
	return x
}

func (e *News) Delete(userID ulid.ULID, topic string, url ...string) {

	if e.Find(userID, topic) == nil {
		return
	}

	if len(url) == 0 {
		s3.DeletePrivateObject(e.key())
		return
	}

	if a, ok := e.Articles[url[0]]; ok {
		new(Page).Delete(a.URL, a.ID)
		delete(e.Articles, a.URL)
		e.Save()
	}
}

func (e *News) Find(userID ulid.ULID, topic string) *News {

	e.userID = userID
	e.Topic = topic

	if out, err := s3.GetPrivateObject(e.key()); err != nil {
		return nil
	} else if err = json.Unmarshal(out, e); err != nil {
		return nil
	}

	return e
}

func (e *News) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"articles": slices.SortedFunc(maps.Values(e.Articles), func(a, z model.Article) int {
			return a.PublishedAt.Compare(z.PublishedAt)
		}),
		"topic": e.Topic,
	})
}

func (e *News) UnmarshalJSON(b []byte) error {

	var alias struct {
		Topic    string          `json:"topic"`
		Articles []model.Article `json:"articles"`
	}

	if err := json.Unmarshal(b, &alias); err != nil {
		return err
	}

	e.Topic = alias.Topic

	e.Articles = make(map[string]model.Article)
	for _, a := range alias.Articles {
		e.Articles[a.URL] = a
	}

	return nil
}
