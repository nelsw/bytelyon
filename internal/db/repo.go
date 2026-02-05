package db

import "gorm.io/gorm"

func FindAllEnabled[T any](db *gorm.DB, t *[]T) error {
	return db.Scopes(Enabled).Find(t).Error
}
