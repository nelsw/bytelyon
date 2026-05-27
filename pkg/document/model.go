package document

import (
	"maps"
	"regexp"
	"slices"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/nelsw/bytelyon/pkg/meta"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/urls"
	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/rs/zerolog/log"
)

type Model struct {
	doc    *goquery.Document
	domain string
	Meta   meta.Model
	url    string
}

func New(url, content string) *Model {

	m := &Model{
		domain: urls.Domain(url),
		url:    url,
		Meta:   make(meta.Model),
	}

	var err error
	if m.doc, err = goquery.NewDocumentFromReader(strings.NewReader(content)); err != nil {
		log.Warn().Err(err).Msg("failed to parse html")
		m.doc = &goquery.Document{}
		return m
	}

	m.doc.Find("meta").Each(func(i int, s *goquery.Selection) {
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

func (m *Model) Links() []string {
	x := model.NewSet[string]()
	m.doc.Find("a").Each(func(i int, s *goquery.Selection) {

		href := strings.TrimSpace(s.AttrOr("href", ""))

		// trim whitespace (yes, technically it's possible)
		if href = strings.TrimSpace(href); href == "" {
			return
		}

		// is it a js link?
		if strings.Contains(href, "javascript:") || strings.Contains(href, "about:blank") {
			return
		}

		// is it a file link?
		if urls.HasFileExtension(href) {
			return
		}

		// is it a fragment?
		if strings.HasPrefix(href, "#") {
			return
		}

		// is it a browser function?
		if regexp.MustCompile(`^(mailto|tel|sms|fax|callto|geo):.*`).MatchString(href) {
			return
		}

		x.Add(href)
	})
	return x.Slice()
}

func (m *Model) Title() string { return util.Or(m.Meta.Title(), m.doc.Find("title").Text()) }

func (m *Model) Headings() map[string][]string {
	x := make(map[string][]string)
	for _, h := range []string{"h1", "h2", "h3", "h4", "h5", "h6"} {
		m.doc.Find(h).Each(func(i int, s *goquery.Selection) {
			x[h] = append(x[h], strings.TrimSpace(s.Text()))
		})
	}
	return x
}

func (m *Model) Paragraphs() []string {
	var x []string
	uniqueParagraphs := make(map[string]int)
	m.doc.Find("p").Each(func(i int, s *goquery.Selection) {
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

func (m *Model) URLs() []string {
	set := model.NewSet[string]()
	for _, link := range m.Links() {
		// if the link is an insecure URL
		if strings.HasPrefix(link, "http://") {
			continue
		}

		// if the link is empty or root
		if link == "" || link == "/" {
			continue
		}

		// if the link is relative to the root urls
		if strings.HasPrefix(link, "/") {
			set.Add("https://" + m.domain + link)
			continue
		}

		// if the link is a url; check the host equals our domain
		if host := urls.Host(link); host != "" && host != m.domain {
			continue
		}

		// if the link is a secure URL
		if strings.HasPrefix(link, "https://"+m.domain) {
			set.Add(link)
			continue
		}

		// if the link is missing URL protocol
		if strings.HasPrefix(link, m.domain) {
			set.Add("https://" + link)
			continue
		}

		// else the link is relative to this url
		if l, _, ok := strings.Cut(link, "/"); ok {
			set.Add(m.url + "/" + l + "/" + link)
		} else {
			set.Add(m.url + "/" + link)
		}
	}
	return set.Slice()
}
