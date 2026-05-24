package entity

import (
	"fmt"

	"github.com/nelsw/bytelyon/pkg/em"
	"github.com/nelsw/bytelyon/pkg/https"
	"github.com/oklog/ulid/v2"
)

type PageScreenshot struct {
	id   ulid.ULID
	data []byte
	url  string
}

func (p *PageScreenshot) Associations() []em.Entity { return nil }
func (p *PageScreenshot) Val() []byte               { return p.data }
func (p *PageScreenshot) Key() string {
	return fmt.Sprintf("page/%s/%s.png", https.Trim(p.url), p.id)
}
