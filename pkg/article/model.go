package article

import (
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/oklog/ulid/v2"
)

type Model struct {
	Body        []string    `json:"body"`
	Description string      `json:"description"`
	ID          ulid.ULID   `json:"id"`
	Image       model.Image `json:"image"`
	Keywords    []string    `json:"keywords"`
	Meta        model.Meta  `json:"meta"`
	Source      string      `json:"source"`
	Title       string      `json:"title"`
	URL         string      `json:"url"`
}
