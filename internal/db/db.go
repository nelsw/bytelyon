package db

import (
	"database/sql"
	"time"

	"github.com/nelsw/bytelyon/internal/config"
	"github.com/nelsw/bytelyon/internal/model"
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
}

func Init() {

	var err error
	if db, err = gorm.Open(sqlite.Open(config.Mode()+".sqlite"), &gorm.Config{}); err != nil {
		panic(err)
	}

	var sqlDB *sql.DB
	if sqlDB, err = db.DB(); err != nil || sqlDB.Ping() != nil {
		panic(err)
	}

	sqlDB.SetMaxIdleConns(25)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(time.Hour * 24)

	if config.IsTestMode() {
		for _, t := range models {
			if err = db.Migrator().DropTable(&t); err != nil {
				panic(err)
			}
		}
	}

	if err = db.AutoMigrate(models...); err != nil {
		panic(err)
	}
}

func Builder[T any]() gorm.Interface[T] {
	return gorm.G[T](db)
}
