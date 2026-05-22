package model

import (
	"encoding/json"
	"fmt"

	"github.com/nelsw/bytelyon/pkg/s3"
	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

type Sitemap struct {
	Domain string

	*SyncMap[string, []ulid.ULID]

	UserID ulid.ULID
}

func NewSitemap(userID ulid.ULID, domain string) *Sitemap {
	return &Sitemap{
		Domain:  domain,
		UserID:  userID,
		SyncMap: NewSyncMap[string, []ulid.ULID](),
	}
}

func (e *Sitemap) From(userID ulid.ULID, domain string) *Sitemap {
	if x := e.Find(userID, domain); x != nil {
		return x
	}
	return NewSitemap(userID, domain).Save()
}

func (e *Sitemap) key() string { return fmt.Sprintf("users/%s/sitemaps/%s.json", e.UserID, e.Domain) }

func (e *Sitemap) Save() *Sitemap {
	s3.PutPrivateObject(e.key(), util.JSON(e))
	return e
}

func (e *Sitemap) Create(userID ulid.ULID, domain string) *Sitemap {
	x := &Sitemap{
		Domain:  domain,
		SyncMap: NewSyncMap[string, []ulid.ULID](),
		UserID:  userID,
	}
	x.Save()
	return x
}

func (e *Sitemap) Delete(userID ulid.ULID, domain string) {
	if e.Find(userID, domain) != nil {
		s3.DeletePrivateObject(e.key())
	}
}

func (e *Sitemap) Find(userID ulid.ULID, domain string) *Sitemap {

	e.UserID = userID
	e.Domain = domain

	if out, err := s3.GetPrivateObject(e.key()); err != nil {
		return nil
	} else if err = json.Unmarshal(out, e); err != nil {
		return nil
	}

	log.Trace().Str("domain", e.Domain).Msg("sitemap found")

	return e
}

func (e *Sitemap) Add(p *Page) {

	if p == nil {
		return
	}

	ids, ok := e.Get(p.URL)
	if !ok {
		ids = []ulid.ULID{}
	}

	e.Set(p.URL, append(ids, p.ID))
	e.Save()
}

func (e *Sitemap) URL() string { return "https://" + e.Domain }

func (e *Sitemap) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.ToMap())
}

func (e *Sitemap) UnmarshalJSON(b []byte) error {
	var m map[string][]ulid.ULID
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	e.SyncMap = NewSyncMap[string, []ulid.ULID](m)
	return nil
}
