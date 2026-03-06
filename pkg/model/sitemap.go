package model

import (
	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/oklog/ulid/v2"
)

type Sitemap struct {
	ID     ulid.ULID
	UserID ulid.ULID
	Target string
	Links  map[URL]bool
}

func (b Sitemap) String() string  { return "sitemap" }
func (b Sitemap) Table() *string  { return util.Ptr("Sitemap_Bot") }
func (b Sitemap) Validate() error { return nil }
