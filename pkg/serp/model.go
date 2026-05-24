package serp

import (
	"strings"

	"github.com/nelsw/bytelyon/internal/pw"
	"github.com/nelsw/bytelyon/pkg/id"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/oklog/ulid/v2"
	"github.com/playwright-community/playwright-go"
)

type Model struct {
	ID ulid.ULID `json:"-"`

	URL string `json:"url"`

	// Entries is a map of SERP sections to SERP section results.
	Entries map[string][]any `json:"entries"`

	content    string
	screenshot []byte
}

func New(page playwright.Page, q string) *Model {
	m := &Model{
		ID:         id.New(),
		URL:        "google.com/search?q=" + strings.ReplaceAll(q, " ", "+"),
		Entries:    make(map[string][]any),
		content:    pw.Content(page),
		screenshot: pw.Screenshot(page),
	}
	for k, result := range model.MakeSerp(m.URL, m.content) {
		m.Entries[string(k)] = append(m.Entries[string(k)], result)
	}
	return m
}
