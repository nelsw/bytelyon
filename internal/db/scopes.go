package db

import "gorm.io/gorm"

func Enabled(db *gorm.DB) *gorm.DB {
	return db.Where("enabled = ?", true)
}
