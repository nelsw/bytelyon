package model

type SearchPage struct {
	Model
	SearchID uint
	Search   *Search `json:",omitempty"`
	URL      string
	Title    string
	IMG      string
	HTML     string
	JSON     any `gorm:"serializer:json"`
}
