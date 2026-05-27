package news

import (
	"github.com/oklog/ulid/v2"
)

type Model struct {
	ID    ulid.ULID `json:"id"`
	Title string    `json:"title"`
	URL   string    `json:"url"`
}
