package rss

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

	"github.com/nelsw/bytelyon/pkg/https"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/html"
)

var (
	gStaticRegex = regexp.MustCompile(`/articles/(?P<encoded_url>[^?]+)`)
)

type Item struct {
	Title       string    `json:"title"`
	Link        string    `json:"link"`
	Description string    `json:"description,omitempty"`
	PublishedAt time.Time `json:"publishedAt"`
	Source      string    `json:"source,omitempty"`
}

func Items(url string, after time.Time, exclude map[string]bool) []*Item {
	var res []*Item

	b, err := https.Get(url)
	if err != nil {
		log.Err(err).Str("url", url).Msg("Failed to get RSS feed")
		return res
	}

	for _, item := range strings.Split(string(b), "</item>") {

		_, str, ok := strings.Cut(item, "<item>")
		if !ok {
			continue
		}

		var i Item

		/*
			pubDate
		*/
		i.PublishedAt, _ = time.Parse(time.RFC1123, getBetween(str, "<pubDate>", "</pubDate>"))
		if i.PublishedAt.IsZero() {
			i.PublishedAt = time.Now()
		} else if i.PublishedAt.Before(after) {
			continue
		}

		/*
			link
		*/
		i.Link = getBetween(str, "<link>", "</link>")
		if strings.Contains(url, "news.google.com") {
			i.Link = decodeGoogleURL(i.Link)
		} else if strings.Contains(url, "bing.com") {
			i.Link = decodeBingURL(i.Link)
		}
		if exclude[i.Link] {
			continue
		}

		/*
			description
		*/
		if !strings.Contains(url, "google.com") {
			i.Description = getBetween(str, "<description>", "</description>")
		}

		/*
			title
		*/
		i.Title = getBetween(str, "<title>", "</title>")

		/*
			source
		*/
		if strings.Contains(str, "<News:Source>") {
			i.Source = getBetween(str, "<News:Source>", "</News:Source>")
		} else {
			i.Source = getBetween(str, `<source`, "</source>")
			if _, r, k := strings.Cut(i.Source, `>`); k {
				i.Source = r
			}
		}

		res = append(res, &i)
	}

	return res
}

func getBetween(str, start, end string) string {
	s := strings.Index(str, start)
	if s == -1 {
		return ""
	}
	s += len(start)
	e := strings.Index(str[s:], end)
	if e == -1 {
		return ""
	}
	return str[s : s+e]
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

func decodeGoogleURL(s string) string {

	b, err := https.Get(s)
	if err != nil {
		log.Warn().Str("url", s).Msg("failed to get gstatic html")
		return s
	}

	var doc *html.Node
	if doc, err = html.Parse(bytes.NewReader(b)); err != nil {
		log.Warn().Str("url", s).Msg("failed to parse gstatic html")
		return s
	}

	matches := gStaticRegex.FindStringSubmatch(s)
	if len(matches) < 2 {
		log.Warn().Str("url", s).Msg("failed to match gstatic regex")
		return s
	}

	var out string
	if out, err = decodeNode(doc, matches[1]); err != nil {
		log.Warn().Str("url", s).Msg("failed to decode gstatic node")
		return s
	}

	log.Debug().Str("url", s).Str("out", out).Msg("decoded gstatic url")
	return out
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
	payload := []any{
		"Fbv4je",
		fmt.Sprintf("[\"garturlreq\",[[\"X\",\"X\",[\"X\",\"X\"],null,null,1,1,\"US:en\",null,1,null,null,null,null,null,0,1],\"X\",\"X\",1,[1,1,1],1,1,null,0,0,null,0],\"%s\",%s,\"%s\"]", base64Str, timestamp, signature),
	}
	outer := [][]any{payload}
	bodyBytes, _ := json.Marshal([][][]any{outer})
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

	payload = []any{}
	if err = json.Unmarshal([]byte(parts[1]), &payload); err != nil {
		return "", err
	} else if len(payload) == 0 {
		return "", errors.New("empty payload")
	}

	entry, ok := payload[0].([]any)
	if !ok || len(entry) < 3 {
		return "", errors.New("unexpected entry structure")
	}

	var inner []any
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
