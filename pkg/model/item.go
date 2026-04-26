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

	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/html"
)

var (
	gStaticRegex = regexp.MustCompile(`/articles/(?P<encoded_url>[^?]+)`)
)

type Item struct {
	URL         string `xml:"link"`
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Source      string `xml:"source"`
	PublishedAt *Time  `xml:"pubDate"`
	NewsSource  string `xml:"News_Source"`
}

func (i *Item) String() string {
	return fmt.Sprintf("{\n"+
		"\turl: %s,\n"+
		"\ttitle: %s,\n"+
		"\tdescription: %s,\n"+
		"\tsource: %s,\n"+
		"\ttime: %s,\n"+
		"}",
		i.URL,
		i.Title,
		i.Description,
		i.Source,
		i.PublishedAt,
	)
}

func (i *Item) ContainsKeyword(keyword string) bool {

	log.Info().Str("keyword", keyword).Msg("checking article for keyword")

	if strings.Contains(strings.ToLower(i.Title), keyword) {
		log.Info().
			Str("keyword", keyword).
			Str("title", i.Title).
			Msg("article title contains keyword")
		return true
	}

	if strings.Contains(strings.ToLower(i.Description), keyword) {
		log.Info().
			Str("keyword", keyword).
			Str("description", i.Description).
			Msg("article description contains keyword")
		return true
	}

	log.Info().Str("keyword", keyword).Msg("article does not contain keyword")
	return false
}

func (i *Item) ProcessXML() {

	if strings.Contains(i.URL, "bing.com") {
		i.URL = decodeBingURL(i.URL)
		return
	}

	if strings.Contains(i.URL, "google.com") {
		s, err := decodeGoogleURL(i.URL)
		if err != nil {
			log.Warn().Err(err).Msg("failed to decode google")
			return
		}
		i.URL = s
	}

	// populate a potentially empty source
	if i.Source == "" {
		i.Source = i.NewsSource
		if i.Source == "" {
			i.Source = util.Domain(i.URL)
		}
	}

	// remove source from the title if it exists
	i.Title = strings.TrimPrefix(i.Title, " - "+i.Source)

	// remove html from the description
	if idx := strings.Index(i.Description, `</a>`); idx > 0 {
		i.Description = i.Description[:idx]
		i.Description = i.Description[strings.LastIndex(i.Description, ">")+1:]
	}
}

func (i *Item) IsBlacklisted(args []string) bool {
	for _, arg := range args {
		if i.ContainsKeyword(arg) {
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
