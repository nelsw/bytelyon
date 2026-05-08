package model

import (
	"time"

	"github.com/oklog/ulid/v2"
)

type Post struct {
	ID          ulid.ULID `json:"id"`
	Handle      string    `json:"handle"`
	Title       string    `json:"title,omitempty"`
	Body        string    `json:"body,omitempty"`
	Summary     string    `json:"summary,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
	ImgSrc      string    `json:"imgSrc,omitempty"`
	ImgAlt      string    `json:"imgAlt,omitempty"`
	PublishedAt time.Time `json:"publishedAt"`
}
