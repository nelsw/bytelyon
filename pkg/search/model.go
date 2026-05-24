package search

import (
	"fmt"

	"github.com/oklog/ulid/v2"
)

type Model struct {
	Entries map[ulid.ULID][]string `json:"entries"`

	Query string `json:"-"`

	UserID ulid.ULID `json:"-"`
}

func New(userID ulid.ULID, query string) *Model {
	return &Model{
		Query:   query,
		UserID:  userID,
		Entries: make(map[ulid.ULID][]string),
	}
}

func (m *Model) Key() string {
	return fmt.Sprintf("users/%s/search/%s.json", m.UserID, m.Query)
}

func (m *Model) Merge(other *Model) {
	for k, v := range other.Entries {
		m.Entries[k] = append(m.Entries[k], v...)
	}
}
