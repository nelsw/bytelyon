package model

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/html"
)

type SerpSection string

const (
	SponsoredSection           SerpSection = "sponsored"
	OrganicSection             SerpSection = "organic"
	VideoSection               SerpSection = "video"
	ForumSection               SerpSection = "forum"
	ArticleSection             SerpSection = "article"
	PopularProductsSection     SerpSection = "popular_products"
	MoreProductsSection        SerpSection = "more_products"
	PeopleAlsoAskSection       SerpSection = "people_also_ask"
	PeopleAlsoSearchForSection SerpSection = "people_also_search_for"
)

type SerpResult struct {
	Position int     `json:"position"`
	Title    string  `json:"title"`
	Link     string  `json:"link"`
	Source   string  `json:"source"`
	Snippet  string  `json:"snippet"`
	Price    float64 `json:"price,omitempty"`
}

type Serp map[SerpSection][]SerpResult

func MakeSerp(url, content string) Serp {

	if !strings.HasPrefix(url, "https://www.google.com") {
		return nil
	}

	serp := make(Serp)

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err != nil {
		log.Warn().Err(err).Msg("Page - failed to parse html")
		return nil
	}

	serp.fillSponsoredData(doc, content)

	serp.fillOrganicData(content)
	serp.fillOrganicDataV2(doc)

	serp.fillPeopleAlsoAskData(doc)
	serp.fillPeopleAlsoAskDataV2(doc)

	serp.fillPeopleAlsoSearchForData(doc)
	serp.fillPeopleAlsoSearchForDataV2(doc)

	return serp
}

func (s Serp) fillSponsoredData(doc *goquery.Document, content string) {

	var ids []string
	doc.Find(`div`).Each(func(i int, sel *goquery.Selection) {
		if _, ok := sel.Attr("data-merchant-id"); !ok {
			return
		}
		if id, ok := sel.Attr("id"); ok && id[0] == '_' {
			ids = append(ids, id)
		}
	})

	var frags []string
	for _, id := range ids {
		if left := strings.Index(content, id+`',`); left > 0 {
			left += len(id) + 4
			right := strings.Index(content[left:], `);})();`)
			frags = append(frags, content[left:left+right])
		}
	}

	for i := range frags {
		frags[i] = strings.ReplaceAll(frags[i], "x26", "&")
		frags[i] = strings.ReplaceAll(frags[i], "x27", "'")
		frags[i] = strings.ReplaceAll(frags[i], "xb2", "²")
		frags[i] = strings.ReplaceAll(frags[i], "x3d", "=")
		frags[i] = strings.ReplaceAll(frags[i], "x22", "")
		frags[i] = strings.ReplaceAll(frags[i], "x3c", "<")
		frags[i] = strings.ReplaceAll(frags[i], "x3e", ">")
		frags[i] = strings.ReplaceAll(frags[i], "&amp;", "&")
		frags[i] = strings.ReplaceAll(frags[i], `\`, ``)
	}

	for pos, f := range frags {
		var datum = SerpResult{}
		datum.Position = pos

		d, err := html.Parse(strings.NewReader(f))
		if err != nil {
			log.Warn().Err(err).Msg("failed to parse sponsored html")
			continue
		}

		gd := goquery.NewDocumentFromNode(d)

		goquery.NewDocumentFromNode(d).Find(`span`).Each(func(i int, sel *goquery.Selection) {

			t := strings.TrimSpace(sel.Text())
			if len(t) == 0 || t[0] != '$' {
				return
			}

			t = strings.ReplaceAll(t, ",", "")[1:]
			if price, e := strconv.ParseFloat(t, 64); e == nil {
				datum.Price = price
			}
		})

		gd.Find(`div`).Each(func(i int, sel *goquery.Selection) {

			t := strings.TrimSpace(sel.Text())

			if _, ok := sel.Attr("aria-label"); ok {
				datum.Source = t
				return
			}

			if val, ok := sel.Attr("role"); ok && val == "heading" {
				datum.Title = t
				return
			}

		})

		gd.Find(`a`).Each(func(i int, sel *goquery.Selection) {
			if datum.Link != "" {
				return
			}
			if val, ok := sel.Attr("href"); ok && strings.Contains(val, "https://") {
				datum.Link = val
				return
			}
		})

		s[SponsoredSection] = append(s[SponsoredSection], datum)
	}
}

func (s Serp) fillOrganicData(content string) {
	left := strings.Index(content, `var m={`) + 7
	content = content[left:]
	content = content[:strings.Index(content, "};")]

	var vals []string
	for i, chunk := range strings.Split(content, `:[`) {

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

		var d = SerpResult{}
		d.Position = pos

		val := strings.ReplaceAll(vals[pos], "null,", "")
		for i, v := range strings.Split(val, ",[") {

			if i == 0 {
				v = strings.ReplaceAll(v, "\"", "")
				v, _ = strconv.Unquote("\"" + v + "\"")
				d.Link = v[1:]
			} else if i == 2 {
				for i, v = range strings.Split(v, "\",\"") {
					switch i {
					case 0:
						v = strings.ReplaceAll(v, "\\u003d", "=")
						v = strings.ReplaceAll(v, "\\u0026", "&")
						v = strings.ReplaceAll(v, "\\", "")
						v = strings.ReplaceAll(v, "\"", "")
						d.Title = v
					case 1:
						v, _ = strconv.Unquote("\"" + v + "\"")
						d.Snippet = v
					case 2:
						d.Source = v
					}
				}
				break
			}
		}
		if strings.Contains(val, "WEB_RESULT_INNER") {
			d.Position = len(s[OrganicSection])
			s[OrganicSection] = append(s[OrganicSection], d)
		} else if strings.Contains(val, "COMMUNITY_MODE_WEB_RESULT") {
			d.Position = len(s[ForumSection])
			s[ForumSection] = append(s[ForumSection], d)
		} else if strings.Contains(val, "VIDEO_RESULT") {
			d.Position = len(s[VideoSection])
			s[VideoSection] = append(s[VideoSection], d)
		} else if strings.Contains(val, "NEWS_ARTICLE_RESULT") {
			d.Position = len(s[ArticleSection])
			s[ArticleSection] = append(s[ArticleSection], d)
		}
	}
}
func (s Serp) fillOrganicDataV2(doc *goquery.Document) []SerpResult {

	e := doc.Find("a").FilterFunction(func(i int, s *goquery.Selection) bool {
		href, ok := s.Attr("href")
		return ok &&
			!strings.Contains(href, "google.com") &&
			strings.HasPrefix(href, "/url") &&
			strings.Contains(href, "&url=")
	})

	if e == nil {
		return nil
	}

	var arr []SerpResult
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

		arr = append(arr, SerpResult{
			Position: len(arr),
			Link:     URL,
			Title:    first,
			Snippet:  second,
			Source:   last,
		})
	}

	return arr
}

func (s Serp) fillPeopleAlsoAskData(doc *goquery.Document) {
	doc.Find("div[class*='related-question-pair']").Each(func(i int, sel *goquery.Selection) {
		if sel == nil {
			return
		}
		s[PeopleAlsoAskSection] = append(s[PeopleAlsoAskSection], SerpResult{
			Position: len(s[PeopleAlsoAskSection]),
			Title:    sel.AttrOr("data-q", ""),
			Source:   "Google",
		})
	})
}
func (s Serp) fillPeopleAlsoAskDataV2(doc *goquery.Document) []SerpResult {
	e := doc.Find("span:contains('People also ask')")
	if e == nil {
		return nil
	}

	var i int
	for i == 0 {
		i = e.Siblings().Size()
		e = e.Parent()
	}

	e = e.Find("div").FilterFunction(func(i int, s *goquery.Selection) bool {
		return strings.HasSuffix(s.Text(), "?")
	})
	if e == nil {
		return nil
	}

	x := make(map[string]bool)
	m := make(map[int]string)
	for _, e = range e.EachIter() {
		if _, ok := x[e.Text()]; ok {
			continue
		}
		x[e.Text()] = true
		m[len(m)] = e.Text()
	}

	d := make([]SerpResult, 0, len(m))
	for i = 0; i < len(m)-1; i++ {
		d[i] = SerpResult{Position: i, Title: m[i]}
	}
	return d
}

func (s Serp) fillPeopleAlsoSearchForData(doc *goquery.Document) {
	doc.Find("span").Each(func(i int, sel *goquery.Selection) {
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

			s[PeopleAlsoSearchForSection] = append(s[PeopleAlsoSearchForSection], SerpResult{
				Position: len(s[PeopleAlsoSearchForSection]) + 1,
				Link:     fmt.Sprintf("https://www.google.com%s", sel.AttrOr("href", "")),
				Title:    sel.Text(),
				Source:   "Google",
			})
		})
	})
}
func (s Serp) fillPeopleAlsoSearchForDataV2(doc *goquery.Document) []SerpResult {
	e := doc.Find("accordion-entry-search-icon")
	if e == nil {
		return nil
	}

	x := make(map[string]bool)
	m := make(map[int]string)
	for _, e = range e.EachIter() {
		if e.Next() == nil {
			continue
		}
		txt := e.Next().Text()
		if _, ok := x[txt]; ok {
			continue
		}
		x[txt] = true
		m[len(m)] = txt
	}

	d := make([]SerpResult, 0, len(m))
	for i := 0; i < len(m)-1; i++ {
		d[i] = SerpResult{Position: i + 1, Title: m[i]}
	}
	return d
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

// products: <product-viewer-entrypoint
// ads: href="/aclk?
// role="heading"
// <g-card
