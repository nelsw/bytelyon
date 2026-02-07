package db

import (
	"github.com/nelsw/bytelyon/internal/model"
)

var Migrations = []any{
	&model.Bot{},
	&model.Article{},
	&model.Sitemap{},
	&model.Search{},
	&model.SearchPage{},
}
