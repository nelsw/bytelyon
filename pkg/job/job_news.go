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

	"github.com/PuerkitoBio/goquery"
	"github.com/nelsw/bytelyon/pkg/client"
	"github.com/nelsw/bytelyon/pkg/db"
	"github.com/rs/zerolog"
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
func (v *Time) Before(t time.Time) bool { return v.UTC().Before(t.UTC()) }
func (v *Time) UTC() time.Time          { return time.Time(*v).UTC() }

type RSS struct {
	Channel struct {
		Items []*Item `xml:"item"`
	} `xml:"channel"`
}

type Item struct {
	URL         string `xml:"link"`
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Source      string `xml:"source"`
	Time        *Time  `xml:"pubDate"`
	NewsSource  string `xml:"News_Source"`
	Image       string `xml:"-"`
	Content     string `xml:"-"`
}

func (i *Item) MarshalZerologObject(evt *zerolog.Event) {
	evt.Str("url", i.URL).
		Str("title", i.Title).
		Stringer("time", i.Time).
		Str("description", i.Description).
		Str("source", i.Source).
		Str("news_source", i.NewsSource).
		Str("image", i.Image).
		Str("content", i.Content)
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

func (i *Item) ProcessHTML() {

	log.Info().Object("item", i).Msg("processing item html")

	res, err := http.Get(i.URL)
	if err != nil {
		log.Warn().Err(err).Object("item", i).Msg("failed to fetch URL to hydrate news HTML")
		return
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	var b []byte
	if b, err = io.ReadAll(res.Body); err != nil {
		log.Warn().Err(err).Object("item", i).Msg("Failed to read article url bytes")
		return
	}
	i.Content = string(b)

	var doc *goquery.Document
	if doc, err = goquery.NewDocumentFromReader(res.Body); err != nil {
		log.Warn().Err(err).Msg("failed to create doc to hydrate news HTML")
		return
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

	for _, m := range mm {

		// define the title if empty
		if i.Title == "" {
			if v, k := m["property"]; k && v == "og:title" {
				if v, k = m["content"]; k {
					i.Title = v
				}
			}
		}

		// define the description if empty
		if i.Description == "" {
			if v, k := m["name"]; k && v == "description" || v == "twitter:description" {
				if v, k = m["content"]; k {
					i.Description = v
				}
			}
		}
		if i.Description == "" {
			if v, k := m["property"]; k && v == "description" || v == "og:description" {
				if v, k = m["content"]; k {
					i.Description = v
				}
			}
		}

		// define the source if empty
		if i.Source == "" {
			if v, k := m["property"]; k && v == "og:site_name" {
				if v, k = m["content"]; k {
					i.Source = v
				}
			}
		}

		// define the image if empty
		if i.Image == "" {
			if v, k := m["name"]; k && v == "twitter:image" {
				if v, k = m["content"]; k {
					i.Image = v
					log.Debug().Msg("found image to process news HTML: " + v)
				}
			}
		}
		if i.Image == "" {
			if v, k := m["property"]; k && v == "og:image" {
				if v, k = m["content"]; k {
					i.Image = v
					log.Debug().Msg("found image to process news HTML: " + v)
				}
			}
		}
	}

	if i.Title == "" {
		doc.Find("title").Each(func(idx int, s *goquery.Selection) { i.Title = s.Text() })
	}

	log.Info().Msg("processed news HTML")
}

func (j *Job) doNews() {

	log.Info().Msgf("processing news job %s", j.bot.Target)

	q := strings.ReplaceAll(j.bot.Target, ` `, `+`)
	urls := []string{
		fmt.Sprintf("https://news.google.com/rss/search?q=%s&hl=en-US&gl=US&ceid=US:en", q),
		fmt.Sprintf("https://www.bing.com/news/search?format=rss&q=%s", q),
		fmt.Sprintf("https://www.bing.com/search?format=rss&q=%s", q),
	}

	var wg sync.WaitGroup
	for _, u := range urls {
		wg.Go(func() { j.doNewsFeed(u) })
	}
	wg.Wait()

	log.Info().Msgf("processing news job %s", j.bot.Target)
}

func (j *Job) doNewsFeed(u string) {
	res, err := http.Get(u)
	if err != nil {
		log.Err(err).Str("url", u).Msg("Failed to fetch RSS feed")
		return
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	var b []byte
	if b, err = io.ReadAll(res.Body); err != nil {
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
		wg.Go(func() { j.doNewsFeedArticle(i) })
	}
	wg.Wait()
}

func (j *Job) doNewsFeedArticle(i *Item) {

	log.Info().Object("item", i).Msg("Processing RSS item (news article)")

	// if this job is brand new, save all the articles found
	// else persist articles published after the last update
	if !j.bot.WorkedAt.IsZero() && i.Time.Before(j.bot.WorkedAt) {
		log.Info().
			Stringer("published", i.Time).
			Time("worked", j.bot.WorkedAt.UTC()).
			Msgf("Skipping old article %s", i.Title)
		return
	}

	// work some magic to circumvent protected urls
	i.URL = decodeURL(i.URL)

	// include html data
	i.ProcessHTML()

	// try to populate an empty source
	if i.Source == "" {
		i.Source = i.NewsSource
	}

	// remove the source if it exists
	i.Title = strings.TrimSuffix(i.Title, " - "+i.Source)

	// check if the description is HTML
	if idx := strings.Index(i.Description, `</a>`); idx > 0 {
		i.Description = i.Description[:idx]
		i.Description = i.Description[strings.LastIndex(i.Description, ">")+1:]
	}

	// now that the item is populated with data,
	// check article for blacklisted keywords
	for kw := range j.rules {
		if i.ContainsKeyword(kw) {
			return
		}
	}

	// instantiate a new bot result
	result := j.bot.NewBotResult(
		"url", i.URL,
		"title", i.Title,
		"source", i.Source,
		"description", i.Description,
		"publishedAt", i.Time.UTC(),
	)

	// save article html if it exits and define the path on the result
	if i.Content != "" {
		// define the s3 bucket key for article html
		key := fmt.Sprintf("users/%s/bots/news/%s/content/%s.html",
			j.bot.UserID,
			j.bot.Target,
			result.ID,
		)

		if err := client.PutObject(j.ctx, j.s3, "bytelyon-public", key, []byte(i.Content)); err != nil {
			log.Warn().Err(err).Object("item", i).Msg("Failed to save news article html")
		} else {
			result.Data["content"] = key
			i.Content = key
		}
	}

	// save article image if it exists and define the path on the result
	if i.Image != "" {

		res, err := http.Get(i.Image)
		if err != nil {
			log.Warn().Err(err).Object("item", i).Msg("Failed to download news article")
		} else {
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(res.Body)

			var b []byte
			if b, err = io.ReadAll(res.Body); err != nil {
				log.Warn().Err(err).Object("item", i).Msg("Failed to read news article")
			} else if idx := strings.LastIndex(i.Image, "."); idx > 0 {

				ext := i.Image[idx+1:]
				if idx = strings.LastIndex(i.Image, "?"); idx > 0 {
					ext = ext[:idx]
				}

				// define the s3 bucket key for article image
				key := fmt.Sprintf("users/%s/bots/news/%s/image/%s.%s",
					j.bot.UserID,
					j.bot.Target,
					result.ID,
					ext,
				)

				if err = client.PutObject(j.ctx, j.s3, "bytelyon-public", key, b); err != nil {
					log.Warn().Err(err).Object("item", i).Msg("Failed to save news article image")
				} else {
					result.Data["image"] = key
					i.Image = key
				}
			}
		}
	}

	// save the result
	if err := db.PutItem(result); err != nil {
		log.Warn().Err(err).Object("item", i).Msg("Failed to save news article")
	} else {
		log.Info().Object("item", i).Msg("News article saved")
	}
}

func decodeURL(s string) string {

	if strings.Contains(s, "bing.com") {
		return decodeBingURL(s)
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
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	var doc *html.Node
	if doc, err = html.Parse(res.Body); err != nil {
		return s, err
	}

	matches := regex.FindStringSubmatch(s)
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
