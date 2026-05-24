package article

import (
	"github.com/nelsw/bytelyon/internal/pw"
	"github.com/nelsw/bytelyon/pkg/document"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/oklog/ulid/v2"
	"github.com/playwright-community/playwright-go"
)

type Models []*Model

func (m Models) Len() int           { return len(m) }
func (m Models) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m Models) Less(i, j int) bool { return m[i].ID.Compare(m[j].ID) < 0 }

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

	content    string
	screenshot []byte
}

func (m *Model) Fill(ctx playwright.BrowserContext) {
	m.content, m.screenshot = pw.Scrape(m.URL, ctx)
	doc := document.New(m.content)
	if m.Description == "" {
		m.Description = doc.Description()
	}
	if len(m.Image) == 0 {
		m.Image = doc.Image()
	}
	if m.Image.GetSrc() == "" {
		m.Image.SetSrc(doc.ImageSrc())
	}
	if m.Image.GetAlt() == "" {
		m.Image.SetAlt(doc.ImageAlt())
	}
	m.Keywords = doc.Keywords()
	m.Meta = doc.Meta
	m.Body = doc.Paragraphs
}
