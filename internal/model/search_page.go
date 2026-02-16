package model

import "gorm.io/gorm"

type SearchPage struct {
	gorm.Model
	Search   *Search
	SearchID uint
	URL      string
	Title    string
	IMG      string
	HTML     string
	JSON     any `gorm:"serializer:json"`
}
