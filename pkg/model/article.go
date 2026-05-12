package model

import (
	"strings"
	"time"
)

type Article struct {
	URL           string    `json:"url"`
	Title         string    `json:"title"`
	Body          []string  `json:"body"`
	Image         Image     `json:"image"`
	PublishedAt   time.Time `json:"publishedAt"`
	Source        string    `json:"source"`
	Description   string    `json:"description"`
	Keywords      []string  `json:"keywords"`
	ScreenshotURL string    `json:"screenshotUrl,omitempty"`
}

func (a Article) Words() []string {

	allTxt := strings.Join(a.Keywords, " ") +
		" " + strings.Join(a.Body, " ") +
		" " + a.Description +
		" " + a.Title

	return strings.Fields(allTxt)
}
