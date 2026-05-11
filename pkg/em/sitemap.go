package em

import (
	"github.com/nelsw/bytelyon/pkg/entity"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/store"
	"github.com/oklog/ulid/v2"
)

func sitemapDB(userID ulid.ULID) (*store.DB[string, *entity.Sitemap], error) {
	return store.New[string, *entity.Sitemap]("users", userID, "bots", model.SitemapBotType)
}

func GetSitemap(userID ulid.ULID, domain string) (*entity.Sitemap, bool) {
	db, err := sitemapDB(userID)
	if err != nil {
		return nil, false
	}
	defer db.Close()
	return db.Get(domain)
}

func SaveSitemap(userID ulid.ULID, e *entity.Sitemap) {

	db, err := sitemapDB(userID)
	if err != nil {
		return
	}
	defer db.Close()

	if x, ok := db.Get(e.Domain); ok {
		e.AddURLs(x.URLs)
	}
	db.Set(e.Domain, e)
}

func DeleteSitemap(userID ulid.ULID, domain string) {
	db, err := sitemapDB(userID)
	if err != nil {
		return
	}
	defer db.Close()
	db.Drop(domain)
}
