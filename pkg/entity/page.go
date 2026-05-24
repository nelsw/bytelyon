package entity

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/nelsw/bytelyon/internal/pw"
	"github.com/nelsw/bytelyon/pkg/https"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/s3"
	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/oklog/ulid/v2"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

type Page struct {
	ID            ulid.ULID           `json:"id"`
	URL           string              `json:"url"`
	Title         string              `json:"title"`
	Meta          model.Meta          `json:"meta"`
	Domain        string              `json:"domain"`
	Path          string              `json:"path"`
	Links         []string            `json:"links"`
	Headings      map[string][]string `json:"headings"`
	Paragraphs    []string            `json:"paragraphs"`
	Image         model.Image         `json:"image"`
	PublishedAt   time.Time           `json:"publishedAt"`
	Source        string              `json:"source"`
	Description   string              `json:"description"`
	Keywords      []string            `json:"keywords"`
	SERP          model.Serp          `json:"serp"`
	ScreenshotURL string              `json:"screenshotUrl"`
	Screenshot    []byte              `json:"-"`
	Content       string              `json:"-"`
}

func FindPage(url string, id ulid.ULID) *Page {
	e := &Page{URL: url, ID: id}
	if out, err := s3.GetPrivateObject(e.key("json")); err != nil {
		return nil
	} else if err = json.Unmarshal(out, e); err != nil {
		return nil
	}

	return e
}

func NewPage(p playwright.Page) *Page {

	url := strings.TrimSuffix(p.URL(), "/")
	content := pw.Content(p)
	meta := model.MakeMeta(pw.Meta(p))

	return &Page{
		Content:     content,
		Description: meta.Desc(),
		Domain:      util.Domain(url),
		Headings:    pw.Headings(p),
		ID:          model.NewULID(),
		Image:       meta.Img(),
		Keywords:    meta.Keywerds(),
		Links:       pw.Links(p),
		Meta:        meta,
		Paragraphs:  pw.Paragraphs(p),
		PublishedAt: meta.PublishedAt(),
		Screenshot:  pw.Screenshot(p),
		SERP:        model.MakeSerp(url, content),
		Source:      meta.Source(),
		Title:       util.Or(pw.Title(p), meta.Titel()),
		URL:         url,
	}
}

func (p *Page) key(ext string) string {
	return fmt.Sprintf("page/%s/%s.%s", https.Trim(p.URL), p.ID, ext)
}

func (p *Page) Save() {

	if len(p.Screenshot) > 0 {
		key := p.key("png")
		s3.PutPublicImage(key, p.Screenshot)
		p.ScreenshotURL = "https://bytelyon-public.s3.amazonaws.com/" + key
	}

	s3.PutPrivateObject(p.key("json"), util.JSON(p))

	if p.SERP != nil {
		s3.PutPrivateObject(p.key("html"), []byte(p.Content))
	}
}

func (p *Page) Scrape(url string, ctx playwright.BrowserContext) *Page {

	l := log.With().
		Str("ƒ", "scrape").
		Str("url", url).
		Logger()

	l.Trace().Send()

	page, err := pw.NewPage(ctx)
	if err != nil {
		l.Warn().Msgf("scrape failed: %s", err.Error())
		return nil
	}

	defer page.Close()

	if err = pw.Visit(page, url); err != nil {
		l.Warn().Msgf("Visit failed: %s", err.Error())
		return nil
	}

	l.Debug().Send()

	return NewPage(page)
}

func (p *Page) Delete(url string, id ulid.ULID) {
	p.URL = url
	p.ID = id
	s3.DeletePrivateObject(p.key(".json"))
	s3.DeletePublicImage(p.key(".png"))
}

func (p *Page) MakeSnippet() Snippet {
	return Snippet{
		ID:            p.ID,
		URL:           p.URL,
		Domain:        util.Domain(p.URL),
		Title:         p.Title,
		Meta:          p.Meta,
		ScreenshotURL: p.ScreenshotURL,
	}
}

func (p *Page) Words() []string {

	allTxt := strings.Join(p.Keywords, " ") +
		" " + strings.Join(p.Paragraphs, " ") +
		" " + p.Description +
		" " + p.Title

	return strings.Fields(allTxt)
}
