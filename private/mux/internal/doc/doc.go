package doc

import (
	"bytes"
	"maps"
	"slices"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/nelsw/bytelyon/apps/mux/internal/util/gzip"
)

type Meta map[string][]string

type Doc struct {
	*goquery.Document
	body        *string
	description *string
	imgAlt      *string
	imgUrl      *string
	keywords    []string
	meta        Meta
	source      *string
	title       *string
}

func New(src []byte, compressed bool) (doc *Doc) {

	var err error
	if compressed {
		if src, err = gzip.Decompress(src); err != nil {
			return
		}
	}

	doc = new(Doc)
	doc.Document, _ = goquery.NewDocumentFromReader(bytes.NewReader(src))

	return
}

func (d *Doc) value(keys ...string) string {
	var opts []string
	for _, key := range keys {
		opts = append(opts, d.Meta()[key]...)
	}
	for _, opt := range opts {
		if opt = strings.TrimSpace(opt); opt != "" {
			return opt
		}
	}
	return ""
}

func (d *Doc) Meta() Meta {
	if d.meta != nil {
		return d.meta
	}
	d.meta = make(Meta)
	d.Find("meta").Each(func(idx int, s *goquery.Selection) {
		k := s.AttrOr("name", s.AttrOr("property", ""))
		v := s.AttrOr("content", "")

		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)

		if k == "" || v == "" {
			return
		}

		m := make(map[string]bool)
		for _, v = range strings.Split(v, ",") {
			if v = strings.TrimSpace(v); v != "" {
				m[v] = true
			}
		}
		d.meta[k] = slices.Sorted(maps.Keys(m))
	})
	return d.meta
}

func (d *Doc) MetaValue(key string) []string {
	return d.Meta()[key]
}

func (d *Doc) Title() string {
	if d.title == nil {
		val := d.value("twitter:title", "og:title", "title")
		d.title = &val
	}
	return *d.title
}

func (d *Doc) Keywords() []string {

	if len(d.keywords) > 0 {
		return d.keywords
	}

	var opts []string
	opts = append(opts, d.Meta()["keywords"]...)
	opts = append(opts, d.Meta()["news_keywords"]...)
	opts = append(opts, d.Meta()["article:tag"]...)

	m := make(map[string]bool)
	for _, opt := range opts {
		m[opt] = true
	}

	d.keywords = slices.Sorted(maps.Keys(m))

	return d.keywords
}

func (d *Doc) Source() string {
	if d.source == nil {
		val := d.value("twitter:site", "og:site_name", "og:site")
		d.source = &val
	}
	return *d.source
}

func (d *Doc) Description() string {
	if d.description == nil {
		val := d.value("twitter:description", "og:description", "description", "abstract")
		d.description = &val
	}
	return *d.description
}

func (d *Doc) ImgAlt() string {
	if d.imgAlt == nil {
		val := d.value("twitter:image:alt", "og:image:alt")
		d.imgAlt = &val
	}
	return *d.imgAlt
}

func (d *Doc) ImgUrl() string {
	if d.imgUrl == nil {
		val := d.value("twitter:image:src", "twitter:image", "og:image:secure_url", "og:image:url", "og:image", "image")
		d.imgUrl = &val
	}
	return *d.imgUrl
}

func (d *Doc) Body() string {
	if d.body != nil {
		return *d.body
	}
	sel := d.Find("article")
	if len(sel.Nodes) == 0 {
		sel = d.Find("main")
		if len(sel.Nodes) == 0 {
			sel = d.Find("body")
			if len(sel.Nodes) == 0 {
				sel = d.Find("html")
			}
		}
	}
	var text []string
	sel.Contents().Each(func(i int, el *goquery.Selection) {
		if txt := strings.TrimSpace(el.Text()); txt != "" {
			text = append(text, txt)
		}
	})

	body := strings.Join(text, "\n\n")
	d.body = &body

	return body
}
