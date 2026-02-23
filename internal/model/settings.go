package model

import "gorm.io/gorm"

type Settings struct {
	gorm.Model
	AWS
}
