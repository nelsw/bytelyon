package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/nelsw/bytelyon/internal/config"
	"github.com/nelsw/bytelyon/internal/model"
	"github.com/rs/zerolog/log"
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
	if db, err = gorm.Open(sqlite.Open(config.Mode()+".db.sqlite"), &gorm.Config{}); err != nil {
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

func Create(a any) (err error) {
	if err = db.Create(&a).Error; err != nil {
		log.Err(err).Send()
	}
	return
}

func Delete[T any](scopes ...func(statement *gorm.Statement)) (err error) {
	if _, err = gorm.G[T](db).Scopes(scopes...).Delete(context.Background()); err != nil {
		log.Err(err).Send()
	}
	return
}

func Find[T any](scopes ...func(*gorm.DB) *gorm.DB) (arr []T, err error) {
	if err = db.Scopes(scopes...).Find(&arr).Error; err != nil {
		log.Err(err).Send()
	}
	return
}

func Save(a any) (err error) {
	if err = db.Save(&a).Error; err != nil {
		log.Err(err).Send()
	}
	return
}

func MustDelete[T any](scopes ...func(statement *gorm.Statement)) {
	if err := Delete[T](scopes...); err != nil {
		panic(err)
	}
}

func MustFind[T any](scopes ...func(*gorm.DB) *gorm.DB) []T {
	arr, err := Find[T](scopes...)
	if err != nil {
		panic(err)
	}
	return arr
}

func MustSave(a any) {
	if err := Save(a); err != nil {
		panic(err)
	}
}
