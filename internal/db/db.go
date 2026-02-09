package db

import (
	"database/sql"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func New(mode string) *gorm.DB {

	db, err := gorm.Open(sqlite.Open(mode+".v2.sqlite"), &gorm.Config{})
	if err != nil {
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

	if mode == gin.TestMode {
		for _, t := range Migrations {
			if err = db.Migrator().DropTable(&t); err != nil {
				panic(err)
			}
		}
	}

	if err = db.AutoMigrate(Migrations...); err != nil {
		panic(err)
	}

	return db
}
