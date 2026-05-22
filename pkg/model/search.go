package model

import (
	"encoding/json"
	"fmt"

	"github.com/nelsw/bytelyon/pkg/s3"
	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/oklog/ulid/v2"
)

type Search struct {
	ID       ulid.ULID `json:"id"`
	Query    string    `json:"query"`
	Snippets []Snippet `json:"snippets"`
	Serp     Serp      `json:"serp"`
	UserID   ulid.ULID `json:"-"`
}

func (e *Search) key() string {
	return fmt.Sprintf("users/%s/searches/%s.json", e.UserID, e.Query)
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

func (e *Search) Find(userID ulid.ULID, query string) *Search {

	e.UserID = userID
	e.Query = query

	if out, err := s3.GetPrivateObject(e.key()); err != nil {
		return nil
	} else if err = json.Unmarshal(out, e); err != nil {
		return nil
	}

	return e
}

func (e *Search) Delete(userID ulid.ULID, query string) {
	if e.Find(userID, query) != nil {
		s3.DeletePrivateObject(e.key())
	}
}

func (e *Search) Create(userID ulid.ULID, query string) *Search {
	x := &Search{
		ID:     NewULID(),
		Query:  query,
		UserID: userID,
	}
	x.Save()
	return x
}
