package model

import "gorm.io/gorm"

type Sitemap struct {
	gorm.Model
	JobID    uint
	Job      Job
	URL      string
	Domain   string
	Relative []string `gorm:"serializer:json"`
	Remote   []string `gorm:"serializer:json"`
}
