package job

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/nelsw/bytelyon/pkg/db"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/html"
)

var (
	bingRegexp = regexp.MustCompile("</?News(:\\w+)>")
	regex      = regexp.MustCompile(`/articles/(?P<encoded_url>[^?]+)`)
)

type Time time.Time

func (v *Time) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var s string
	if err := d.DecodeElement(&s, &start); err != nil {
		return err
	}

	s = strings.Trim(s, `"`) // Remove quotes from the JSON string
	if s == "" || s == "null" {
		return nil // Handle empty or null strings
	}

	t, err := time.Parse(time.RFC1123, s) // Parse using your custom format
	if err != nil {
		return err
	}

	*v = Time(t.UTC())
	return nil
}
func (v *Time) String() string          { return v.UTC().Format(time.RFC3339) }
func (v *Time) Before(t time.Time) bool { return time.Time(*v).UTC().Before(t) }
func (v *Time) UTC() time.Time          { return time.Time(*v).UTC() }

type RSS struct {
	Channel struct {
		Items []*struct {
			URL         string `xml:"link"`
			Title       string `xml:"title"`
			Description string `xml:"description"`
			Source      string `xml:"source"`
			Time        *Time  `xml:"pubDate"`
			NewsSource  string `xml:"News_Source"`
		} `xml:"item"`
	} `xml:"channel"`
}

func (j *Job) doNews() {
	q := strings.ReplaceAll(j.bot.Target, ` `, `+`)
	urls := []string{
		fmt.Sprintf("https://news.google.com/rss/search?q=%s&hl=en-US&gl=US&ceid=US:en", q),
		fmt.Sprintf("https://www.bing.com/news/search?format=rss&q=%s", q),
		fmt.Sprintf("https://www.bing.com/search?format=rss&q=%s", q),
	}
	for _, u := range urls {
		j.doNewsFeed(u)
	}
}

func (j *Job) doNewsFeed(u string) {
	r, err := http.Get(u)
	if err != nil {
		log.Err(err).Str("url", u).Msg("Failed to fetch RSS feed")
		return
	}
	defer r.Body.Close()

	var b []byte
	if b, err = io.ReadAll(r.Body); err != nil {
		log.Err(err).Str("url", u).Msg("Failed to read RSS feed")
		return
	}

	if strings.Contains(u, "bing.com") {
		b = []byte(bingRegexp.ReplaceAllStringFunc(string(b), func(s string) string {
			return strings.ReplaceAll(s, ":", "_")
		}))
	}

	var rss RSS
	if err = xml.Unmarshal(b, &rss); err != nil {
		log.Err(err).Str("url", u).Msg("Failed to unmarshal RSS feed")
		return
	}

	var wg sync.WaitGroup
	for _, i := range rss.Channel.Items {

		wg.Go(func() {

			log.Trace().Any("item", i).Msg("Processing RSS item")

			// if this job is brand new, save all the articles found
			// else persist articles published after the last update
			if !j.bot.WorkedAt.IsZero() && i.Time.Before(j.bot.WorkedAt) {
				log.Info().
					Stringer("published", i.Time).
					Stringer("worked", j.bot.WorkedAt).
					Msgf("Skipping old article %s", i.Title)
				return
			}

			// check article data for blacklisted keywords
			titleParts := strings.Split(i.Title, " ")
			sourceParts := strings.Split(i.Source, " ")
			parts := append(titleParts, sourceParts...)
			for _, p := range parts {
				if follow, exists := j.rules[p]; exists && !follow {
					log.Info().Msgf("Skipping blacklisted article %s", p)
					return
				}
			}

			// work some magic to circumvent protected urls
			i.URL = decodeURL(i.URL)

			// check if the source is blank and use the news source if it is
			if i.Source == "" && i.NewsSource != "" {
				i.Source = i.NewsSource
			}

			// scrub the source off the title and use it if the item source is blank
			if l, r, ok := strings.Cut(i.Title, " - "); ok {
				i.Title = l
				if i.Source == "" {
					i.Source = r
				}
			}

			// check if the description is HTML
			if idx := strings.Index(i.Description, `</a>`); idx > 0 {
				i.Description = i.Description[:idx]
				i.Description = i.Description[strings.LastIndex(i.Description, ">")+1:]
			}

			err = db.PutItem(j.bot.NewBotResult(
				"url", i.URL,
				"title", i.Title,
				"source", i.Source,
				"description", i.Description,
				"publishedAt", i.Time.UTC(),
			))

			log.Err(err).Msg("put news result")
		})
	}
	wg.Wait()
}

func decodeURL(s string) string {

	if strings.Contains(s, "bing.com") {
		s = decodeBingURL(s)
	} else if strings.Contains(s, "google.com") {
		var err error
		if s, err = decodeGoogleURL(s); err != nil {
			log.Warn().Err(err).Msg("failed to decode Google URL")
		}
	}

	return s
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
	defer res.Body.Close()

	var doc *html.Node
	if doc, err = html.Parse(res.Body); err != nil {
		return s, err
	}

	return decodeNode(doc, regex.FindStringSubmatch(s)[1])
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

	client := &http.Client{}
	var resp *http.Response
	if resp, err = client.Do(req); err != nil {
		return "", err
	}
	defer resp.Body.Close()
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
