package document

import (
	"maps"
	"regexp"
	"slices"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/rs/zerolog/log"
)

type Model struct {
	Title string

	Headings map[string][]string

	Links []string

	Paragraphs []string

	model.Meta
}

func New(content string) (m *Model) {

	m = new(Model)

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err != nil {
		log.Warn().Err(err).Msg("failed to parse html")
		return
	}

	/*
		Title
	*/
	m.Title = doc.Find("title").Text()

	/*
		Headings
	*/
	m.Headings = make(map[string][]string)
	for _, h := range []string{"h1", "h2", "h3", "h4", "h5", "h6"} {
		doc.Find(h).Each(func(i int, s *goquery.Selection) {
			m.Headings[h] = append(m.Headings[h], strings.TrimSpace(s.Text()))
		})
	}

	/*
		Links
	*/
	var links = make(map[string]bool)
	doc.Find("a").Each(func(i int, s *goquery.Selection) {

		href := strings.TrimSpace(s.AttrOr("href", ""))

		// trim whitespace (yes, technically it's possible)
		if href = strings.TrimSpace(href); href == "" {
			return
		}

		// is it a js link?
		if strings.Contains(href, "javascript:") {
			return
		}

		// is it a file link?
		if util.HasFileExtension(href) {
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

		links[href] = true
	})
	m.Links = slices.Collect(maps.Keys(links))

	/*
		Paragraphs
	*/
	uniqueParagraphs := make(map[string]int)
	doc.Find("p").Each(func(i int, s *goquery.Selection) {
		if k := strings.TrimSpace(s.Text()); k != "" {
			uniqueParagraphs[s.Text()] = i
		}
	})
	orderedParagraphs := make(map[int]string)
	for k, v := range uniqueParagraphs {
		orderedParagraphs[v] = k
	}
	for _, k := range slices.Sorted(maps.Keys(orderedParagraphs)) {
		m.Paragraphs = append(m.Paragraphs, orderedParagraphs[k])
	}

	/*
		Meta
	*/
	m.Meta = make(map[string]string)
	doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		if k := strings.ToLower(s.AttrOr("name", s.AttrOr("property", ""))); k == "" {
			return
		} else if v := s.AttrOr("content", ""); v == "" {
			return
		} else {
			m.Meta[k] = v
		}

	})

	return
}
