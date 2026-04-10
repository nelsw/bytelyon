package repo

import (
	"github.com/nelsw/bytelyon/pkg/db"
	"github.com/nelsw/bytelyon/pkg/model"
)

func SaveSitemap(newS *model.Sitemap) error {

	oldS := model.NewSitemap(newS.Domain)

	db.Get(oldS)

	oldS.URLs.Add(newS.URLs)

	return db.Put(oldS)
}
