package model

import "github.com/oklog/ulid/v2"

type Snippet struct {
	ID  ulid.ULID `json:"id"`
	URL string    `json:"url"`

	Domain string `json:"domain"`
	Path   string `json:"path"`

	Title string `json:"title"`
	Meta  Meta   `json:"meta"`

	PageRank    int `json:"pageRank"`
	SectionRank int `json:"sectionRank"`

	ScreenshotURL string `json:"screenshotUrl"`
}
