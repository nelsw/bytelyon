package model

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/html"
)

var (
	gStaticRegex = regexp.MustCompile(`/articles/(?P<encoded_url>[^?]+)`)
)

type Article struct {
	URL         string `xml:"link"`
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Source      string `xml:"source"`
	Time        *Time  `xml:"pubDate"`
	NewsSource  string `xml:"News_Source"`
	Image       string `xml:"-"`
	Content     string `xml:"-"`
	Link        string `xml:"-"`
}

func (a *Article) MarshalZerologObject(evt *zerolog.Event) {
	evt.Str("url", a.URL).
		Str("title", a.Title).
		Stringer("time", a.Time).
		Str("description", a.Description).
		Str("source", a.Source).
		Str("news_source", a.NewsSource).
		Str("image", a.Image).
		Str("content", a.Content)
}

// IsOldNews returns true if the given time is not zero
// and Article publication date is before the given time.
// We use this method to assist in preventing data dupes.
func (a *Article) IsOldNews(t time.Time) bool {
	return !t.IsZero() && !a.Time.Before(t)
}

func (a *Article) ContainsKeyword(keyword string) bool {

	log.Info().Str("keyword", keyword).Msg("checking article for keyword")

	if strings.Contains(strings.ToLower(a.Title), keyword) {
		log.Info().
			Str("keyword", keyword).
			Str("title", a.Title).
			Msg("article title contains keyword")
		return true
	}

	if strings.Contains(strings.ToLower(a.Description), keyword) {
		log.Info().
			Str("keyword", keyword).
			Str("description", a.Description).
			Msg("article description contains keyword")
		return true
	}

	log.Info().Str("keyword", keyword).Msg("article does not contain keyword")
	return false
}

func (a *Article) DecodeURL() {

	if strings.Contains(a.URL, "bing.com") {
		a.URL = decodeBingURL(a.URL)
		return
	}

	if strings.Contains(a.URL, "google.com") {
		s, err := decodeGoogleURL(a.URL)
		if err != nil {
			log.Warn().Err(err).Msg("failed to decode google")
			return
		}
		a.URL = s
	}
}

func (a *Article) ProcessHTML() {

	log.Info().Object("item", a).Msg("processing item html")

	res, err := http.Get(a.URL)
	if err != nil {
		log.Warn().Err(err).Object("item", a).Msg("failed to fetch URL to hydrate news HTML")
		return
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	log.Info().
		Int("status", res.StatusCode).
		Str("url", a.URL).
		Msg("got news url")

	var doc *goquery.Document
	if doc, err = goquery.NewDocumentFromReader(res.Body); err != nil {
		log.Warn().Err(err).Msg("failed to create doc to hydrate news HTML")
		return
	}

	if a.Content, err = doc.Html(); err != nil {
		log.Warn().Err(err).Msg("failed to extract news HTML")
	}

	var mm []map[string]string

	doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		var m = make(map[string]string)
		for _, str := range []string{"name", "property", "itemprop", "content"} {
			if at, ok := s.Attr(str); ok {
				m[str] = at
			}
		}
		if len(m) > 0 {
			mm = append(mm, m)
		}
	})

	log.Info().Any("meta", mm).Msg("processing news HTML meta")

	for _, m := range mm {

		// define the title if empty
		if a.Title == "" {
			if v, k := m["property"]; k && v == "og:title" {
				if v, k = m["content"]; k {
					a.Title = v
				}
			}
		}

		// define the description if empty
		if a.Description == "" {
			if v, k := m["name"]; k && v == "description" || v == "twitter:description" {
				if v, k = m["content"]; k {
					a.Description = v
				}
			}
		}
		if a.Description == "" {
			if v, k := m["property"]; k && v == "description" || v == "og:description" {
				if v, k = m["content"]; k {
					a.Description = v
				}
			}
		}

		// define the source if empty
		if a.Source == "" {
			if v, k := m["property"]; k && v == "og:site_name" {
				if v, k = m["content"]; k {
					a.Source = v
				}
			}
		}

		// define the image if empty
		if a.Image == "" {
			if v, k := m["name"]; k && v == "twitter:image" {
				if v, k = m["content"]; k && util.IsImageFile(v) {
					a.Image = v
				}
			}
		}
		if a.Image == "" {
			if v, k := m["property"]; k && v == "og:image" {
				if v, k = m["content"]; k && util.IsImageFile(v) {
					a.Image = v
				}
			}
		}
	}

	if a.Title == "" {
		doc.Find("title").Each(func(idx int, s *goquery.Selection) { a.Title = s.Text() })
	}

	log.Info().Msg("processed news HTML")
}

func (a *Article) ScrubDetails() {

	// populate a potentially empty source
	if a.Source == "" {
		a.Source = a.NewsSource
		if a.Source == "" {
			a.Source = util.Domain(a.URL)
		}
	}

	// remove source from the title if it exists
	a.Title = strings.TrimPrefix(a.Title, " - "+a.Source)

	// remove html from the description
	if idx := strings.Index(a.Description, `</a>`); idx > 0 {
		a.Description = a.Description[:idx]
		a.Description = a.Description[strings.LastIndex(a.Description, ">")+1:]
	}
}

func (a *Article) IsBlacklisted(args []string) bool {
	for _, arg := range args {
		if a.ContainsKeyword(arg) {
			return true
		}
	}
	return false
}

func decodeBingURL(s string) string {
	s, _ = url.QueryUnescape(s)
	_, r, rOk := strings.Cut(s, "url=")
	if !rOk {
		return s
	}
	l, _, lOk := strings.Cut(r, "&c=")
	if !lOk {
		return r
	}
	return l
}

func decodeGoogleURL(s string) (string, error) {
	res, err := http.Get(s)
	if err != nil {
		return s, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	var doc *html.Node
	if doc, err = html.Parse(res.Body); err != nil {
		return s, err
	}

	matches := gStaticRegex.FindStringSubmatch(s)
	if len(matches) < 2 {
		log.Warn().Str("url", s).Msg("failed to match regex")
		return s, nil
	}

	return decodeNode(doc, matches[1])
}

func decodeNode(n *html.Node, encodedText string) (string, error) {
	if n.Type == html.ElementNode && n.Data == "c-wiz" {

		var sg, ts string
		if e := n.FirstChild; e != nil {
			for _, att := range e.Attr {
				if att.Key == "data-n-a-sg" {
					sg = att.Val
				} else if att.Key == "data-n-a-ts" {
					ts = att.Val
				}
			}
		}
		return decodeParts(sg, ts, encodedText)
	}

	// continue traversing every sibling per child. give em noogies.
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if u, e := decodeNode(c, encodedText); u != "" && e == nil {
			return u, nil
		}
	}
	return "", nil
}

func decodeParts(signature, timestamp, base64Str string) (string, error) {
	endpoint := "https://news.google.com/_/DotsSplashUi/data/batchexecute"
	payload := []interface{}{
		"Fbv4je",
		fmt.Sprintf("[\"garturlreq\",[[\"X\",\"X\",[\"X\",\"X\"],null,null,1,1,\"US:en\",null,1,null,null,null,null,null,0,1],\"X\",\"X\",1,[1,1,1],1,1,null,0,0,null,0],\"%s\",%s,\"%s\"]", base64Str, timestamp, signature),
	}
	outer := [][]interface{}{payload}
	bodyBytes, _ := json.Marshal([][][]interface{}{outer})
	form := url.Values{}
	form.Set("f.req", url.QueryEscape(string(bodyBytes)))

	req, err := http.NewRequest("POST", endpoint, bytes.NewBufferString("f.req="+string(url.QueryEscape(string(bodyBytes)))))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36")

	c := &http.Client{}
	var resp *http.Response
	if resp, err = c.Do(req); err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var b []byte
	if b, err = io.ReadAll(resp.Body); err != nil {
		return "", err
	}

	s := string(b)
	parts := strings.Split(s, "\n\n")
	if len(parts) < 2 {
		return "", errors.New("unexpected batchexecute response format")
	}

	payload = []interface{}{}
	if err = json.Unmarshal([]byte(parts[1]), &payload); err != nil {
		return "", err
	} else if len(payload) == 0 {
		return "", errors.New("empty payload")
	}

	entry, ok := payload[0].([]interface{})
	if !ok || len(entry) < 3 {
		return "", errors.New("unexpected entry structure")
	}

	var inner []interface{}
	if s, ok = entry[2].(string); !ok {
		return "", errors.New("missing inner json string")
	} else if err = json.Unmarshal([]byte(s), &inner); err != nil {
		return "", err
	} else if len(inner) < 2 {
		return "", errors.New("unexpected inner array")
	} else if s, ok = inner[1].(string); !ok {
		return "", errors.New("decoded url not string")
	}

	return s, nil
}
