package model

type Sitemap struct {
	URL      string
	Domain   string
	Relative []string `gorm:"serializer:json"`
	Remote   []string `gorm:"serializer:json"`
}
