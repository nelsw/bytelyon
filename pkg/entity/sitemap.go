package entity

import (
	"encoding/json"
	"fmt"

	"github.com/nelsw/bytelyon/pkg/dto"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/s3"
	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

type Sitemap struct {
	Domain string

	root *dto.Node

	*model.SyncMap[string, []ulid.ULID]

	userID ulid.ULID
}

func (e *Sitemap) From(userID ulid.ULID, domain string) *Sitemap {
	if x := e.Find(userID, domain); x != nil {
		return x
	}
	return e.Create(userID, domain)
}

func (e *Sitemap) key() string { return fmt.Sprintf("users/%s/sitemaps/%s.json", e.userID, e.Domain) }

func (e *Sitemap) Save() { s3.PutPrivateObject(e.key(), util.JSON(e)) }

func (e *Sitemap) Create(userID ulid.ULID, domain string) *Sitemap {
	x := &Sitemap{
		Domain:  domain,
		root:    dto.NewNode(domain, "https://"+domain),
		SyncMap: model.NewSyncMap[string, []ulid.ULID](),
		userID:  userID,
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

	e.userID = userID
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
	e.root.Add(p.URL)
	e.Save()
}

func (e *Sitemap) MarshalJSON() ([]byte, error) {
	var pages int
	for _, ids := range e.Values() {
		pages += len(ids)
	}
	return json.Marshal(map[string]any{
		"domain":    e.Domain,
		"nodes":     []*dto.Node{e.root},
		"data":      e.ToMap(),
		"urlCount":  e.Len(),
		"pageCount": pages,
	})
}

func (e *Sitemap) UnmarshalJSON(b []byte) error {
	var alias struct {
		Domain string `json:"domain"`

		Nodes []*dto.Node `json:"nodes"`

		Map map[string][]ulid.ULID `json:"data"`
	}

	if err := json.Unmarshal(b, &alias); err != nil {
		return err
	}

	e.Domain = alias.Domain

	if len(alias.Nodes) > 0 {
		e.root = alias.Nodes[0]
	}

	e.SyncMap = model.NewSyncMap[string, []ulid.ULID](alias.Map)
	return nil
}
