package model

import "gorm.io/gorm"

type SearchPage struct {
	gorm.Model
	SearchID uint
	Search   Search
	URL      string
	Title    string
	IMG      []byte
	HTML     string
	JSON     any `gorm:"serializer:json"`
}
