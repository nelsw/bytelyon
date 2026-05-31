package document

import (
	"maps"
	"slices"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/nelsw/bytelyon/pkg/image"
	"github.com/rs/zerolog/log"
)

type Model struct {
	*goquery.Document
	Meta map[string]string
}

func New(content string) *Model {

	m := new(Model)
	m.Meta = make(map[string]string)

	var err error
	if m.Document, err = goquery.NewDocumentFromReader(strings.NewReader(content)); err != nil || m.Document == nil {
		log.Warn().Err(err).Msg("failed to parse html")
		m.Document = &goquery.Document{}
		return m
	}

	m.Find("meta").Each(func(i int, s *goquery.Selection) {
		if k := strings.ToLower(s.AttrOr("name", s.AttrOr("property", ""))); k == "" {
			return
		} else if v := s.AttrOr("content", ""); v == "" {
			return
		} else {
			m.Meta[k] = v
		}
	})

	return m
}

func (m *Model) meta(keys ...string) string {
	for _, k := range keys {
		if v, ok := m.Meta[k]; ok {
			return v
		}
	}
	return ""
}

func (m *Model) Title() string {
	if v := m.meta("twitter:title", "og:title", "title"); v != "" {
		return v
	}
	return m.Find("title").Text()
}

func (m *Model) Headings() map[string][]string {
	x := make(map[string][]string)
	for _, h := range []string{"h1", "h2", "h3", "h4", "h5", "h6"} {
		m.Find(h).Each(func(i int, s *goquery.Selection) {
			x[h] = append(x[h], strings.TrimSpace(s.Text()))
		})
	}
	return x
}

func (m *Model) HREFs() []string {
	x := make(map[string]bool)
	m.Find("a").Each(func(i int, s *goquery.Selection) {
		if href, ok := s.Attr("href"); ok {
			x[href] = true
		}
	})
	return slices.Collect(maps.Keys(x))
}

func (m *Model) Paragraphs() []string {
	var x []string
	uniqueParagraphs := make(map[string]int)
	m.Find("p").Each(func(i int, s *goquery.Selection) {
		if k := strings.TrimSpace(s.Text()); k != "" {
			uniqueParagraphs[s.Text()] = i
		}
	})
	orderedParagraphs := make(map[int]string)
	for k, v := range uniqueParagraphs {
		orderedParagraphs[v] = k
	}
	for _, k := range slices.Sorted(maps.Keys(orderedParagraphs)) {
		x = append(x, orderedParagraphs[k])
	}
	return x
}

func (m *Model) Image() *image.Model {
	return &image.Model{
		URL: m.meta("twitter:image:src", "twitter:image", "og:image:secure_url", "og:image:url", "og:image", "image"),
		ALT: m.meta("twitter:image:alt", "og:image:alt"),
	}
}

func (m *Model) Source() string {
	return m.meta("twitter:site", "og:site_name", "og:site")
}

func (m *Model) Description() string {
	return m.meta("twitter:description", "og:description", "description", "abstract")
}

func (m *Model) Keywords() []string {

	opts := []string{
		m.Meta["keywords"],
		m.Meta["news_keywords"],
		m.Meta["article:tag"],
	}

	kw := make(map[string]bool)
	for _, opt := range opts {
		if opt == "" {
			continue
		}
		kws := strings.Split(opt, ",")
		for _, w := range kws {
			kw[strings.TrimSpace(w)] = true
		}
	}

	if len(kw) == 0 {
		return []string{}
	}

	return slices.Sorted(maps.Keys(kw))
}
