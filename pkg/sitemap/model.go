package sitemap

import (
	"fmt"

	"github.com/oklog/ulid/v2"
)

type Model struct {

	// Entries is a map of URLs to the IDs of the pages that are mapped to them.
	Entries map[string][]ulid.ULID `json:"entries"`

	// Domain is the root domain of the site to map.
	Domain string `json:"-"`

	// UserID is the user ID of the user who is requesting the sitemap.
	UserID ulid.ULID `json:"-"`
}

func New(userID ulid.ULID, domain string) *Model {
	return &Model{
		Domain:  domain,
		UserID:  userID,
		Entries: make(map[string][]ulid.ULID),
	}
}

func (m *Model) Merge(other *Model) {
	for k, v := range other.Entries {
		m.Entries[k] = append(m.Entries[k], v...)
	}
}

func (m *Model) Key() string {
	return fmt.Sprintf("users/%s/sitemap/%s.json", m.UserID, m.Domain)
}
