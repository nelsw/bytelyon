package model

type Sitemap struct {
	Model
	Bot      *Bot `json:",omitempty"`
	BotID    uint
	URL      string
	Domain   string
	Relative []string `gorm:"serializer:json"`
	Remote   []string `gorm:"serializer:json"`
}
