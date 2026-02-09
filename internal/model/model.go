package model

import (
	"gorm.io/gorm"
)

type Model struct {
	ID        uint `gorm:"primary_key;"`
	CreatedAt int
	UpdatedAt int
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
