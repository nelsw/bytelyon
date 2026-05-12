package entity

import (
	"fmt"
	"time"

	"github.com/nelsw/bytelyon/internal/pw"
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
	Links         []string            `json:"links"`
	Headings      map[string][]string `json:"headings"`
	Paragraphs    []string            `json:"paragraphs"`
	SERP          model.Serp          `json:"serp,omitempty"`
	ScreenshotURL string              `json:"screenshotUrl,omitempty"`
	Screenshot    []byte              `json:"-"`
	Content       string              `json:"-"`
}

func (p *Page) key(ext string) string {
	return fmt.Sprintf("pages/%s/%s.%s", util.RemoveProtocol(p.URL), p.ID, ext)
}

func (p *Page) Save() {

	if len(p.Screenshot) > 0 {
		key := p.key(".png")
		s3.PutPublicImage(key, p.Screenshot)
		p.ScreenshotURL = "https://bytelyon-public.s3.amazonaws.com/" + key
	}

	s3.PutPrivateObject(p.key(".json"), util.JSON(p))

	if p.SERP != nil {
		s3.PutPrivateObject(p.key(".html"), []byte(p.Content))
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

func NewPage(p playwright.Page) *Page {
	e := &Page{
		ID:         model.NewULID(),
		URL:        p.URL(),
		Title:      pw.Title(p),
		Screenshot: pw.Screenshot(p),
		Meta:       model.MakeMeta(pw.Meta(p)),
		Links:      pw.Links(p),
		Paragraphs: pw.Paragraphs(p),
		Headings:   pw.Headings(p),
	}
	if util.Domain(p.URL()) == "google.com" {
		e.Content = pw.Content(p)
		e.SERP = model.MakeSerp(p.URL(), e.Content)
	}
	return e
}

func (p *Page) MakeArticle(pubDate time.Time, source, description string) model.Article {
	return model.Article{
		ID:            p.ID,
		URL:           p.URL,
		Title:         p.Title,
		Body:          p.Paragraphs,
		Image:         p.Meta.Img(),
		PublishedAt:   util.Or(pubDate, p.Meta.PublishedAt()),
		Source:        util.Or(source, p.Meta.Source()),
		Description:   util.Or(description, p.Meta.Desc()),
		Keywords:      p.Meta.Keywerds(),
		ScreenshotURL: p.ScreenshotURL,
	}
}

func (p *Page) MakeSnippet() model.Snippet {
	return model.Snippet{
		ID:            p.ID,
		URL:           p.URL,
		Domain:        util.Domain(p.URL),
		Path:          util.Path(p.URL),
		Title:         p.Title,
		Meta:          p.Meta,
		ScreenshotURL: p.ScreenshotURL,
	}
}
