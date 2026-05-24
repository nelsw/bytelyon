package entity

import (
	"strings"
	"time"

	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/oklog/ulid/v2"
)

type Article struct {
	Body          []string    `json:"body"`
	Description   string      `json:"description"`
	ID            ulid.ULID   `json:"id"`
	Image         model.Image `json:"image"`
	Keywords      []string    `json:"keywords"`
	PublishedAt   time.Time   `json:"publishedAt"`
	ScreenshotURL string      `json:"screenshotUrl,omitempty"`
	Source        string      `json:"source"`
	Title         string      `json:"title"`
	URL           string      `json:"url"`
}

func (a Article) Words() []string {

	allTxt := strings.Join(a.Keywords, " ") +
		" " + strings.Join(a.Body, " ") +
		" " + a.Description +
		" " + a.Title

	return strings.Fields(allTxt)
}
