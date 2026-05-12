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
}

func NewPage(p playwright.Page) *Page {
	return &Page{
		ID:         model.NewULID(),
		URL:        p.URL(),
		Title:      pw.Title(p),
		Screenshot: pw.Screenshot(p),
		Meta:       model.MakeMeta(pw.Meta(p)),
		Links:      pw.Links(p),
		Paragraphs: pw.Paragraphs(p),
		Headings:   pw.Headings(p),
		SERP:       model.MakeSerp(p.URL(), pw.Content(p)),
	}
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
