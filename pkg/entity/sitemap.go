package entity

import (
	"encoding/json"
	"fmt"

	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/oklog/ulid/v2"
)

type Sitemap struct {
	*model.SyncMap[string, []ulid.ULID]

	Domain string

	UserID ulid.ULID
}

func NewSitemap(userID ulid.ULID, domain string) *Sitemap {
	return &Sitemap{
		Domain:  domain,
		UserID:  userID,
		SyncMap: model.NewSyncMap[string, []ulid.ULID](),
	}
}

func (e *Sitemap) Key() string { return fmt.Sprintf("users/%s/sitemaps/%s.json", e.UserID, e.Domain) }

func (e *Sitemap) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.Keys())
}

func (e *Sitemap) UnmarshalJSON(b []byte) error {
	var m map[string][]ulid.ULID
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	e.SyncMap = model.NewSyncMap[string, []ulid.ULID](m)
	return nil
}
