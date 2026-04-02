package model

import (
	"encoding/json"
	"maps"
	"slices"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/oklog/ulid/v2"
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
	ID                ulid.ULID         `json:"id"`
	Meta              map[string]string `json:"meta"`
	Paragraphs        []string          `json:"paragraphs"`
	Title             string            `json:"title"`
	*goquery.Document `json:"-"`
}

func (d *Document) MarshalZerologObject(evt *zerolog.Event) {
	evt.Stringer("id", d.ID).
		Str("title", d.Title).
		Int("metaCount", len(d.Meta)).
		Int("paragraphCount", len(d.Paragraphs))
}

func (d *Document) String() string {
	b, _ := json.MarshalIndent(d, "", "\t")
	return string(b)
}

func (d *Document) Keywords() []string {
	var m = make(map[string]bool)
	for _, key := range keywordsMetaKeys {
		if csv, ok := d.Meta[key]; ok {
			for _, k := range strings.Split(csv, ",") {
				if strings.Contains(k, ";") {
					continue
				}
				k = strings.TrimSpace(k)
				if k == "" {
					continue
				}
				k = strings.ToLower(k)
				m[k] = true
			}
		}
	}
	return slices.Collect(maps.Keys(m))
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

func (d *Document) setParagraphs() {
	var ss []string
	m := make(map[string]int)
	var txt string
	d.Find("p").Each(func(i int, s *goquery.Selection) {
		if txt = strings.TrimSpace(s.Text()); txt == "" {
			return
		}
		ss = append(ss, txt)
		if _, ok := m[txt]; !ok {
			m[txt] = 0
		} else {
			m[txt] = m[txt] + 1
		}
	})
	for _, s := range ss {
		if m[s] == 0 {
			d.Paragraphs = append(d.Paragraphs, s)
		}
	}
	log.Info().EmbedObject(d).Msg("set paragraphs")
}

func (d *Document) setTitle() {
	d.Find("title").Each(func(i int, s *goquery.Selection) {
		if i == 0 {
			d.Title = s.Text()
			return
		}
	})
	log.Info().EmbedObject(d).Msg("set title")
}

func (d *Document) setMeta() {
	d.Meta = make(map[string]string)
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
}

func NewDocument(id ulid.ULID, content string) (*Document, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	d := &Document{ID: id, Document: doc}
	log.Err(err).EmbedObject(d).Msg("NewDocument")
	if err != nil {
		return nil, err
	}
	d.setMeta()
	d.setTitle()
	d.setParagraphs()
	return d, nil
}
