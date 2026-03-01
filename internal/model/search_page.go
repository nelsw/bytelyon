package model

import "gorm.io/gorm"

type SearchPage struct {
	gorm.Model
	SearchID uint
	URL      string
	Title    string
	IMG      string
	HTML     string
	JSON     any `gorm:"serializer:json"`
}
