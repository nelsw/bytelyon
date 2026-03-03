package model

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/html"
)

type DatumType string

const (
	SponsoredDatumType           DatumType = "sponsored"
	OrganicDatumType             DatumType = "organic"
	VideoDatumType               DatumType = "video"
	ForumDatumType               DatumType = "forum"
	ArticleDatumType             DatumType = "article"
	PopularProductsDatumType     DatumType = "popular_products"
	MoreProductsDatumType        DatumType = "more_products"
	PeopleAlsoAskDatumType       DatumType = "people_also_ask"
	PeopleAlsoSearchForDatumType DatumType = "people_also_search_for"
)

type Datum struct {
	Position int     `json:"position"`
	Title    string  `json:"title"`
	Link     string  `json:"link"`
	Source   string  `json:"source"`
	Snippet  string  `json:"snippet"`
	Price    float64 `json:"price,omitempty"`
}

type PageData struct {
	IDX   int    `json:"idx"`
	URL   string `json:"url"`
	Title string `json:"title"`
	IMG   string `json:"img"`
	HTML  string `json:"html"`
	JSON  any    `json:"json"`
}

func (p *PageData) Parse(content string) {
	if !strings.HasPrefix(p.URL, "https://www.google.com") {
		return
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err != nil {
		log.Warn().Err(err).Msg("Page - failed to parse html")
		return
	}
	m := map[DatumType][]*Datum{
		SponsoredDatumType:           {},
		OrganicDatumType:             {},
		VideoDatumType:               {},
		ForumDatumType:               {},
		ArticleDatumType:             {},
		PopularProductsDatumType:     {},
		MoreProductsDatumType:        {},
		PeopleAlsoAskDatumType:       {},
		PeopleAlsoSearchForDatumType: {},
	}

	fillSponsoredData(doc, content, m)
	fillOrganicData(content, m)
	fillPeopleAlsoAskData(doc, m)
	fillPeopleAlsoSearchForData(doc, m)
}

func fillOrganicData(content string, m map[DatumType][]*Datum) {
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

		var d = new(Datum)
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
			d.Position = len(m[OrganicDatumType])
			m[OrganicDatumType] = append(m[OrganicDatumType], d)
		} else if strings.Contains(val, "COMMUNITY_MODE_WEB_RESULT") {
			d.Position = len(m[ForumDatumType])
			m[ForumDatumType] = append(m[ForumDatumType], d)
		} else if strings.Contains(val, "VIDEO_RESULT") {
			d.Position = len(m[VideoDatumType])
			m[VideoDatumType] = append(m[VideoDatumType], d)
		} else if strings.Contains(val, "NEWS_ARTICLE_RESULT") {
			d.Position = len(m[ArticleDatumType])
			m[ArticleDatumType] = append(m[ArticleDatumType], d)
		}
	}
}

func fillSponsoredData(doc *goquery.Document, content string, m map[DatumType][]*Datum) {

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
		var datum = new(Datum)
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

		m[SponsoredDatumType] = append(m[SponsoredDatumType], datum)
	}
}

func fillPeopleAlsoAskData(doc *goquery.Document, m map[DatumType][]*Datum) {
	doc.Find("div[class*='related-question-pair']").Each(func(i int, sel *goquery.Selection) {
		if sel == nil {
			return
		}
		m[PeopleAlsoAskDatumType] = append(m[PeopleAlsoAskDatumType], &Datum{
			Position: len(m[PeopleAlsoAskDatumType]) + 1,
			Title:    sel.AttrOr("data-q", ""),
			Source:   "Google",
		})
	})
}

func fillPeopleAlsoSearchForData(doc *goquery.Document, m map[DatumType][]*Datum) {
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

			m[PeopleAlsoSearchForDatumType] = append(m[PeopleAlsoSearchForDatumType], &Datum{
				Position: len(m[PeopleAlsoSearchForDatumType]) + 1,
				Link:     fmt.Sprintf("https://www.google.com%s", sel.AttrOr("href", "")),
				Title:    sel.Text(),
				Source:   "Google",
			})
		})
	})
}

// products: <product-viewer-entrypoint
// ads: href="/aclk?
// role="heading"
// <g-card
