package serp

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/nelsw/bytelyon/pkg/document"
	"github.com/nelsw/bytelyon/pkg/id"
	"github.com/nelsw/bytelyon/pkg/util/json"
	"github.com/nelsw/bytelyon/pkg/util/urls"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

type Section string

const (
	article             Section = "article"
	forum               Section = "forum"
	moreProducts        Section = "more_products"
	organic             Section = "organic"
	peopleAlsoAsk       Section = "people_also_ask"
	peopleAlsoSearchFor Section = "people_also_search_for"
	popularProducts     Section = "popular_products"
	sponsored           Section = "sponsored"
	video               Section = "video"
)

func (s Section) Int() int {
	switch s {
	case sponsored:
		return 0
	case organic:
		return 1
	case popularProducts:
		return 2
	case moreProducts:
		return 3
	case video:
		return 4
	case article:
		return 5
	case forum:
		return 6
	case peopleAlsoAsk:
		return 7
	case peopleAlsoSearchFor:
		return 8
	}
	return -1
}

func (s Section) Compare(other Section) int {
	return s.Int() - other.Int()
}

type Items []*Item

func (i Items) Len() int      { return len(i) }
func (i Items) Swap(a, z int) { i[a], i[z] = i[z], i[a] }
func (i Items) Less(a, z int) bool {
	if n := i[a].Section.Int() - i[z].Section.Int(); n != 0 {
		return n < 0
	}
	return i[a].Position < i[z].Position
}

type Item struct {
	Section  Section `json:"section"`
	Link     string  `json:"link"`
	Title    string  `json:"title"`
	Snippet  string  `json:"snippet"`
	Source   string  `json:"source"`
	Position int     `json:"position"`
}

type Model struct {
	content string

	doc *document.Model

	sections map[Section]Items

	ID ulid.ULID `json:"-"`
}

func New(content string) *Model {

	m := &Model{
		ID:       id.NewULID(),
		content:  content,
		doc:      document.New(content),
		sections: make(map[Section]Items),
	}

	m.fillOrganicData()

	m.fillPeopleAlsoAskData()
	m.fillPeopleAlsoAskDataV2()

	m.fillPeopleAlsoSearchForData()
	m.fillPeopleAlsoSearchForDataV2()

	return m
}

func (m *Model) MarshalJSON() ([]byte, error) {
	var items Items
	for _, v := range m.sections {
		items = append(items, v...)
	}
	sort.Sort(items)
	return json.Of(items), nil
}

func (m *Model) UnmarshalJSON(b []byte) error {
	items, err := json.Deserialize[Items](b)
	if err != nil {
		return err
	}

	m.sections = make(map[Section]Items)
	for _, i := range items {
		m.sections[i.Section] = append(m.sections[i.Section], i)
	}
	return nil
}

// Add an item to the models section=>items map
func (m *Model) Add(i *Item) {
	if i == nil {
		log.Warn().Msg("nil item")
		return
	}
	if i.Section == "" {
		log.Warn().Msg("empty category")
		return
	}
	i.Position = len(m.sections[i.Section])

	if (i.Source == "" || strings.Contains(i.Source, " › ")) &&
		(i.Section == sponsored || i.Section == organic) {
		i.Source = urls.Domain(i.Link)
	}
	m.sections[i.Section] = append(m.sections[i.Section], i)
}

func (m *Model) AddSponsored(url, content string) {
	doc := document.New(content)
	m.Add(&Item{
		Section: sponsored,
		Link:    url,
		Title:   doc.Title(),
		Snippet: doc.Description(),
		Source:  doc.Source(),
	})
}

func (m *Model) fillOrganicData() {
	log.Info().Msg("Parsing organic results")
	c := m.content
	left := strings.Index(c, `var m={`) + 7
	c = c[left:]
	c = c[:strings.Index(c, "};")]

	var vals []string
	for i, chunk := range strings.Split(c, `:[`) {

		if i == 0 {
			continue
		}

		idx := strings.LastIndex(chunk, `,`)
		if idx == -1 {
			continue
		}

		key := chunk[idx+1:]
		val := `[` + chunk[:idx]
		_, err := strconv.Atoi(key[len(key)-2 : len(key)-1])

		if err == nil && strings.Contains(val, "Source: ") {
			vals = append(vals, val)
		}
	}

	for pos := 0; pos < len(vals); pos++ {

		item := &Item{}

		val := strings.ReplaceAll(vals[pos], "null,", "")
		for i, v := range strings.Split(val, ",[") {

			if i == 0 {
				v = strings.ReplaceAll(v, "\"", "")
				v, _ = strconv.Unquote("\"" + v + "\"")
				item.Link = v[1:]
			} else if i == 2 {
				for i, v = range strings.Split(v, "\",\"") {
					switch i {
					case 0:
						v = strings.ReplaceAll(v, "\\u003d", "=")
						v = strings.ReplaceAll(v, "\\u0026", "&")
						v = strings.ReplaceAll(v, "\\", "")
						v = strings.ReplaceAll(v, "\"", "")
						item.Title = v
					case 1:
						v, _ = strconv.Unquote("\"" + v + "\"")
						item.Snippet = v
					case 2:
						item.Source = v
					}
				}
				break
			}
		}
		if strings.Contains(val, "WEB_RESULT_INNER") {
			item.Section = organic
		} else if strings.Contains(val, "COMMUNITY_MODE_WEB_RESULT") {
			item.Section = forum
		} else if strings.Contains(val, "VIDEO_RESULT") {
			item.Section = video
		} else if strings.Contains(val, "NEWS_ARTICLE_RESULT") {
			item.Section = article
		}
		m.Add(item)
	}
}

func (m *Model) fillPeopleAlsoAskData() {
	log.Info().Msg("Parsing People Also Ask results")
	m.doc.Find("div[class*='related-question-pair']").Each(func(i int, sel *goquery.Selection) {
		if sel == nil {
			return
		}
		m.Add(&Item{
			Section: peopleAlsoAsk,
			Title:   sel.AttrOr("data-q", ""),
			Source:  "Google",
		})
	})
}
func (m *Model) fillPeopleAlsoAskDataV2() {
	log.Info().Msg("Parsing People Also Ask results v2")

	e := m.doc.Find("span:contains('People also ask')")
	if e == nil {
		return
	}
	var cnt int
	var i int
	for i == 0 {
		i = e.Siblings().Size()
		e = e.Parent()
		cnt++
		if cnt > 100 {
			return
		}
	}

	e = e.Find("div").FilterFunction(func(i int, s *goquery.Selection) bool {
		return strings.HasSuffix(s.Text(), "?")
	})
	if e == nil {
		return
	}

	x := make(map[string]bool)
	mm := make(map[int]string)
	for _, e = range e.EachIter() {
		if _, ok := x[e.Text()]; ok {
			continue
		}
		x[e.Text()] = true
		mm[len(mm)] = e.Text()
	}

	for i = 0; i < len(mm); i++ {
		m.Add(&Item{
			Section: peopleAlsoAsk,
			Title:   mm[i],
		})
	}
}

func (m *Model) fillPeopleAlsoSearchForData() {
	log.Info().Msg("Parsing People Also Search For results")
	m.doc.Find("span").Each(func(i int, sel *goquery.Selection) {
		if sel == nil || sel.Text() != "People also search for" {
			return
		}

		log.Trace().Msg("found people also search for span")

		parent := sel.Parent()
		if parent == nil {
			log.Trace().Msg("no parent found")
			return
		}

		next := parent.Next()
		if next == nil {
			log.Trace().Msg("no next found")
			return
		}

		next.Find("a").Each(func(i int, sel *goquery.Selection) {
			m.Add(&Item{
				Title:  sel.Text(),
				Link:   fmt.Sprintf("https://www.google.com%s", sel.AttrOr("href", "")),
				Source: "Google",
			})
		})
	})
}
func (m *Model) fillPeopleAlsoSearchForDataV2() {
	log.Info().Msg("Parsing People Also Search For results v2")
	e := m.doc.Find("accordion-entry-search-icon")
	if e == nil {
		return
	}

	x := make(map[string]bool)
	y := make(map[int]string)
	for _, e = range e.EachIter() {
		if e.Next() == nil {
			continue
		}
		txt := e.Next().Text()
		if _, ok := x[txt]; ok {
			continue
		}
		x[txt] = true
		y[len(y)] = txt
	}

	for i := 0; i < len(y); i++ {
		m.Add(&Item{
			Section: peopleAlsoSearchFor,
			Title:   y[i],
		})
	}
}
