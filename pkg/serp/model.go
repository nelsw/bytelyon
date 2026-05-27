package serp

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/nelsw/bytelyon/pkg/document"
	"github.com/nelsw/bytelyon/pkg/id"
	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/html"
)

type Model struct {
	Sponsored []any `json:"sponsored"`
	Organic   []any `json:"organic"`
	Forum     []any `json:"forum"`
	Video     []any `json:"video"`
	Article   []any `json:"article"`
	Asked     []any `json:"people_also_ask"`
	Searched  []any `json:"people_also_search_for"`

	ID         ulid.ULID         `json:"-"`
	Doc        *goquery.Document `json:"-"`
	Content    string            `json:"-"`
	URL        string            `json:"-"`
	Screenshot []byte            `json:"-"`
}

func New(query, content string, screenshot []byte) *Model {

	m := &Model{
		ID:         id.NewULID(),
		Doc:        util.Safe(goquery.NewDocumentFromReader(strings.NewReader(content))),
		Content:    content,
		Screenshot: screenshot,
		URL:        "google.com/search?q=" + strings.ReplaceAll(query, " ", "+"),
	}

	m.fillOrganicData()
	if len(m.Organic) == 0 {
		m.fillrOganicDataV2()
	}

	m.fillPeopleAlsoAskData()
	if len(m.Asked) == 0 {
		m.fillPeopleAlsoAskDataV2()
	}

	m.fillPeopleAlsoSearchForData()
	if len(m.Searched) == 0 {
		m.fillPeopleAlsoSearchForDataV2()
	}

	return m
}

func (m *Model) AddSponsored(url, content string) {
	doc := document.New(url, content)
	m.Sponsored = append(m.Sponsored, map[string]any{
		"link":     url,
		"position": len(m.Sponsored),
		"snippet":  doc.Meta.Description(),
		"source":   doc.Meta.Source(),
		"title":    doc.Title(),
	})
}

func (m *Model) fillOrganicData() {
	log.Info().Msg("Parsing organic results")
	c := m.Content
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

		var d = map[string]any{}

		val := strings.ReplaceAll(vals[pos], "null,", "")
		for i, v := range strings.Split(val, ",[") {

			if i == 0 {
				v = strings.ReplaceAll(v, "\"", "")
				v, _ = strconv.Unquote("\"" + v + "\"")
				d["link"] = v[1:]
			} else if i == 2 {
				for i, v = range strings.Split(v, "\",\"") {
					switch i {
					case 0:
						v = strings.ReplaceAll(v, "\\u003d", "=")
						v = strings.ReplaceAll(v, "\\u0026", "&")
						v = strings.ReplaceAll(v, "\\", "")
						v = strings.ReplaceAll(v, "\"", "")
						d["title"] = v
					case 1:
						v, _ = strconv.Unquote("\"" + v + "\"")
						d["snippet"] = v
					case 2:
						d["source"] = v
					}
				}
				break
			}
		}
		if strings.Contains(val, "WEB_RESULT_INNER") {
			d["position"] = len(m.Organic)
			m.Organic = append(m.Organic, d)
		} else if strings.Contains(val, "COMMUNITY_MODE_WEB_RESULT") {
			d["position"] = len(m.Forum)
			m.Forum = append(m.Forum, d)
		} else if strings.Contains(val, "VIDEO_RESULT") {
			d["position"] = len(m.Video)
			m.Video = append(m.Video, d)
		} else if strings.Contains(val, "NEWS_ARTICLE_RESULT") {
			d["position"] = len(m.Article)
			m.Article = append(m.Article, d)
		}
	}
}
func (m *Model) fillrOganicDataV2() {
	log.Info().Msg("Parsing organic results v2")
	e := m.Doc.Find("a").FilterFunction(func(i int, s *goquery.Selection) bool {
		href, ok := s.Attr("href")
		return ok &&
			!strings.Contains(href, "google.com") &&
			strings.HasPrefix(href, "/url") &&
			strings.Contains(href, "&url=")
	})

	if e == nil {
		return
	}

	for _, e = range e.EachIter() {

		l, r, k := strings.Cut(e.AttrOr("href", ""), "&url=")
		if !k {
			continue
		}

		URL := r
		if i := strings.LastIndex(URL, "/"); i > 0 {
			URL = URL[:i]
		}

		for ok := true; ok; _, ok = e.Attr("class") {
			e = e.Parent()
		}

		data, err := e.Html()
		if err != nil {
			continue
		}

		data = strings.ReplaceAll(data, "</a>", "")
		data = strings.ReplaceAll(data, "</div>", "")
		data = strings.ReplaceAll(data, "</span>", "")
		parts := strings.Split(data, "<div>")

		var dollars []string

		for x := 0; x < len(parts); x++ {

			chump := chomp(parts[x])

			for _, chunk := range strings.Split(chump, "<") {

				l, r, k = strings.Cut(chunk, ">")

				var money string
				if !k {
					money = l
				} else {
					money = r
				}

				if money == "" || money == "More results" {
					continue
				}

				money = strings.TrimPrefix(money, " · ")
				money = strings.TrimPrefix(money, "More results")
				money = strings.TrimSuffix(money, "Previous")

				dollars = append(dollars, money)
			}
		}

		if len(dollars) == 0 {
			continue
		}

		var first, second, last string
		for _, d := range dollars {
			d = strings.TrimSpace(d)
			if len(d) < 12 && !strings.Contains(URL, d) {
				continue
			}
			if strings.Contains(d, "›") || strings.Contains(URL, d) {
				last = d
				continue
			}

			d = html.UnescapeString(d)
			if first == "" {
				first = d
				continue
			}
			if second == "" {
				second = d
				continue
			}
			if len(d) > len(second) {
				second = d
			}
		}

		if len(first) > len(second) {
			tmp := first
			first = second
			second = tmp
		}

		m.Organic = append(m.Organic, map[string]any{
			"position": len(m.Organic),
			"link":     URL,
			"title":    first,
			"snippet":  second,
			"source":   last,
		})
	}
}

func (m *Model) fillPeopleAlsoAskData() {
	log.Info().Msg("Parsing People Also Ask results")
	m.Doc.Find("div[class*='related-question-pair']").Each(func(i int, sel *goquery.Selection) {
		if sel == nil {
			return
		}
		m.Asked = append(m.Asked, map[string]any{
			"position": len(m.Asked),
			"title":    sel.AttrOr("data-q", ""),
			"source":   "Google",
		})
	})
}
func (m *Model) fillPeopleAlsoAskDataV2() {
	log.Info().Msg("Parsing People Also Ask results v2")
	e := m.Doc.Find("span:contains('People also ask')")
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
		m.Asked = append(m.Asked, map[string]any{
			"position": len(m.Asked),
			"title":    mm[i],
		})
	}
}

func (m *Model) fillPeopleAlsoSearchForData() {
	log.Info().Msg("Parsing People Also Search For results")
	m.Doc.Find("span").Each(func(i int, sel *goquery.Selection) {
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
			m.Searched = append(m.Searched, map[string]any{
				"position": len(m.Searched) + 1,
				"link":     fmt.Sprintf("https://www.google.com%s", sel.AttrOr("href", "")),
				"title":    sel.Text(),
				"source":   "Google",
			})
		})
	})
}
func (m *Model) fillPeopleAlsoSearchForDataV2() {
	log.Info().Msg("Parsing People Also Search For results v2")
	e := m.Doc.Find("accordion-entry-search-icon")
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
		m.Searched = append(m.Searched, map[string]any{
			"position": len(m.Searched),
			"title":    y[i],
		})
	}
}

func chomp(s string) string {
	for len(s) > 0 && s[0] == '<' {
		idx := strings.Index(s, ">")
		if idx < 0 {
			break
		}
		s = s[idx+1:]
	}
	return s
}
