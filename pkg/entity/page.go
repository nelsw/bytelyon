package entity

import (
	"time"

	"github.com/nelsw/bytelyon/internal/pw"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/oklog/ulid/v2"
	"github.com/playwright-community/playwright-go"
)

type Page struct {
	ID            ulid.ULID           `json:"id"`
	URL           string              `json:"url"`
	Title         string              `json:"title"`
	Meta          map[string]string   `json:"meta"`
	Links         []string            `json:"links"`
	Headings      map[string][]string `json:"headings"`
	Paragraphs    []string            `json:"paragraphs"`
	SERP          any                 `json:"serp,omitempty"`
	ScreenshotURL string              `json:"screenshotUrl,omitempty"`
	Screenshot    []byte              `json:"-"`
	Content       string              `json:"-"`
}

func NewPage(p playwright.Page, t ...time.Time) *Page {
	return &Page{
		ID:         model.NewULID(t...),
		URL:        p.URL(),
		Title:      pw.Title(p),
		Screenshot: pw.Screenshot(p),
		Meta:       pw.Meta(p),
		Links:      pw.Links(p),
		Paragraphs: pw.Paragraphs(p),
		Headings:   pw.Headings(p),
		SERP:       model.MakeSerp(p.URL(), pw.Content(p)),
	}
}

func (p *Page) SetScreenshotURL(url string) {
	p.ScreenshotURL = url
}
