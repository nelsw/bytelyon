package model

import "gorm.io/gorm"

type Search struct {
	gorm.Model
	JobID uint
	Job   Job
	Pages []*SearchPage
}
