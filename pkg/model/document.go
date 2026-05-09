package model

import (
	"encoding/json"
	"errors"
	"maps"
	"slices"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	descriptionMetaKeys = []string{
		"description",
		"og:description",
		"twitter:description",
	}
	titleMetaKeys = []string{
		"twitter:title",
		"og:title",
	}
	imageMetaKeys = []string{
		"og:image",
		"twitter:image:src",
		"twitter:image",
	}
	imageAltMetaKeys = []string{
		"og:image:alt",
		"twitter:image:alt",
	}
	keywordsMetaKeys = []string{
		"keywords",
		"news_keywords",
		"article:section",
	}
	siteMetaKeys = []string{
		"og:site",
		"og:site_name",
		"twitter:site",
	}
)

type Document struct {
	*goquery.Document `json:"-"`

	Content string `json:"content"`

	Title string `json:"title"`

	Meta Data[string] `json:"meta"`

	Paragraphs *Set `json:"paragraphs"`
}

func (d *Document) MarshalZerologObject(evt *zerolog.Event) {
	evt.Str("title", d.Title).
		Int("metaCount", len(d.Meta)).
		Int("paragraphCount", d.Paragraphs.Len())
}

func (d *Document) String() string {
	b, _ := json.MarshalIndent(d, "", "\t")
	return string(b)
}

func (d *Document) MetaSite() (string, bool) {
	return d.meta(siteMetaKeys)
}

func (d *Document) MetaDescription() (string, bool) {
	return d.meta(descriptionMetaKeys)
}

func (d *Document) MetaTitle() (string, bool) {
	return d.meta(titleMetaKeys)
}

func (d *Document) MetaImage() (string, bool) {
	return d.meta(imageMetaKeys, util.IsImageFile)
}

func (d *Document) MetaImageAlt() (string, bool) {
	return d.meta(imageAltMetaKeys)
}

func (d *Document) MetaKeywords() (string, bool) {
	return d.meta(keywordsMetaKeys)
}

func (d *Document) meta(keys []string, ff ...func(s string) bool) (string, bool) {

	var ok bool
	var best, thisStr string
	for _, k := range keys {
		if thisStr, ok = d.Meta[k]; !ok || thisStr == "" {
			continue
		}

		if best != "" && len(best) > len(thisStr) {
			continue
		}

		for _, f := range ff {
			if ok = f(thisStr); !ok {
				break
			}
		}

		if ok {
			best = thisStr
		}
	}

	return best, len(best) > 0
}

func (d *Document) GetParagraphs() *Set {
	if d.Paragraphs == nil || d.Paragraphs.Len() == 0 {
		d.SetParagraphs()
	}
	return d.Paragraphs
}

func (d *Document) GetMeta() Data[string] {
	if d.Meta == nil || d.Meta.Len() == 0 {
		d.SetMeta()
	}
	return d.Meta
}

func (d *Document) SetParagraphs() {

	if d.Paragraphs == nil {
		d.Paragraphs = NewSet()
	}
	if d.Paragraphs.Len() > 0 {
		return
	}
	m := make(map[string]int)
	var txt string
	d.Find("p").Each(func(i int, s *goquery.Selection) {
		if txt = strings.TrimSpace(s.Text()); txt == "" {
			return
		}
		if _, ok := m[txt]; ok {
			return
		}
		m[txt] = i
	})
	log.Info().EmbedObject(d).Msg("set paragraphs")
}

func (d *Document) GetTitle() string {
	if d.Title == "" {
		d.SetTitle()
	}
	if d.Title == "" {
		d.SetMeta()
		d.Title, _ = d.MetaTitle()
	}
	return d.Title
}

func (d *Document) GetImage() Image {
	src, _ := d.MetaImage()
	alt, _ := d.MetaImageAlt()
	return MakeImage(src, alt)
}

func (d *Document) SetTitle() {
	d.Find("title").Each(func(i int, s *goquery.Selection) {
		if i == 0 {
			d.Title = s.Text()
			return
		}
	})
	log.Info().EmbedObject(d).Msg("set title")
}

func (d *Document) SetMeta() {
	if d.Meta == nil {
		d.Meta = MakeData[string]()
	}
	if d.Meta.Len() > 0 {
		return
	}
	var key, val string
	d.Find("meta").Each(func(i int, s *goquery.Selection) {
		key, val = "", ""
		if key = s.AttrOr("name", ""); key == "" {
			if key = s.AttrOr("property", ""); key == "" {
				return
			}
		}
		if val = s.AttrOr("content", ""); val == "" {
			return
		}
		d.Meta[key] = val
	})
	log.Info().EmbedObject(d).Msg("set meta tags")
	log.Trace().Any("meta", d.Meta).Msg("meta tags")
}

func (d *Document) GetHREFs() []string {
	var m = make(map[string]bool)
	d.Find("a").Each(func(i int, s *goquery.Selection) {
		m[s.AttrOr("href", "")] = true
	})
	return slices.Collect(maps.Keys(m))
}

func (d *Document) ToPage(url string, t ...*Time) *Page {
	if len(t) == 0 {
		t = append(t, Now())
	}
	return &Page{
		CreatedAt:   t[0],
		URL:         url,
		Title:       d.GetTitle(),
		Meta:        d.GetMeta(),
		Paragraphs:  d.GetParagraphs(),
		ContentData: d.Content,
	}
}

func ParseDocument(content string) (*Document, error) {

	if content == "" {
		log.Error().Msg("document content is empty")
		return nil, errors.New("document content is empty")
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err != nil {
		log.Err(err).Msg("failed to parse document")
		return nil, err
	}

	return &Document{
		Document:   doc,
		Content:    content,
		Paragraphs: NewSet(),
		Meta:       MakeData[string](),
	}, nil
}
