package news

import (
	"github.com/nelsw/bytelyon/pkg/image"
	"github.com/oklog/ulid/v2"
)

type Article struct {
	Body        []string    `json:"body"`
	Description string      `json:"description"`
	ID          ulid.ULID   `json:"id"`
	Image       image.Model `json:"image"`
	Keywords    []string    `json:"keywords"`
	Source      string      `json:"source"`
	Title       string      `json:"title"`
	URL         string      `json:"url"`
}

type Headline struct {
	ID    ulid.ULID `json:"id"`
	Title string    `json:"title"`
	URL   string    `json:"url,omitempty"`
}
