package db

import (
	"time"

	"github.com/nelsw/bytelyon/internal/config"
	"github.com/nelsw/bytelyon/internal/logger"
	"github.com/nelsw/bytelyon/internal/model"
	. "github.com/nelsw/bytelyon/internal/util"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB
var models = []any{
	&model.Bot{},
	&model.News{},
	&model.Sitemap{},
	&model.Search{},
	&model.SearchPage{},
	&model.Settings{},
}

func Init() {

	db = Must(gorm.Open(sqlite.Open(BinDir(config.Mode()+".sqlite")), &gorm.Config{
		Logger: logger.NewGorm(),
	}))

	sqlDB := Must(db.DB())
	sqlDB.SetMaxIdleConns(25)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(time.Hour * 24)

	if config.IsTestMode() {
		for _, t := range models {
			Check(db.Migrator().DropTable(&t))
		}
	}

	Check(db.AutoMigrate(models...))
}

func Builder[T any]() gorm.Interface[T] {
	return gorm.G[T](db)
}
