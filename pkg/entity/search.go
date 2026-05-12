package entity

import (
	"encoding/json"
	"fmt"

	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/s3"
	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/oklog/ulid/v2"
)

type Search struct {
	ID       ulid.ULID       `json:"id"`
	Query    string          `json:"query"`
	Snippets []model.Snippet `json:"snippets"`
	Serp     model.Serp      `json:"serp"`
	userID   ulid.ULID
}

func (e *Search) key() string {
	return fmt.Sprintf("users/%s/searches/%s.json", e.userID, e.Query)
}

func (e *Search) Save() {
	s3.PutPrivateObject(e.key(), util.JSON(e))
}

func (e *Search) From(userID ulid.ULID, query string) *Search {
	if x := e.Find(userID, query); x != nil {
		return x
	}
	return e.Create(userID, query)
}

func (e *Search) Find(userID ulid.ULID, topic string) *Search {

	e.userID = userID
	e.Query = topic

	if out, err := s3.GetPrivateObject(e.key()); err != nil {
		return nil
	} else if err = json.Unmarshal(out, e); err != nil {
		return nil
	}

	return e
}

func (e *Search) Create(userID ulid.ULID, query string) *Search {
	x := &Search{
		ID:     model.NewULID(),
		Query:  query,
		userID: userID,
	}
	x.Save()
	return x
}
