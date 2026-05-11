package news

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/nelsw/bytelyon/internal/parser"
	"github.com/nelsw/bytelyon/internal/pw"
	"github.com/nelsw/bytelyon/pkg/em"
	"github.com/nelsw/bytelyon/pkg/entity"
	"github.com/nelsw/bytelyon/pkg/https"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/oklog/ulid/v2"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/html"
)

var (
	bingRegexp   = regexp.MustCompile("</?News(:\\w+)>")
	gStaticRegex = regexp.MustCompile(`/articles/(?P<encoded_url>[^?]+)`)
)

type RSS struct {
	Channel struct {
		Items []*Item `xml:"item"`
	} `xml:"channel"`
}

// todo - article
type Item struct {
	URL         string      `xml:"link" json:"url"`
	Title       string      `xml:"title" json:"title"`
	Description string      `xml:"description" json:"summary"`
	Source      string      `xml:"source" json:"source"`
	PublishedAt *model.Time `xml:"pubDate" json:"publishedAt"`
	NewsSource  string      `xml:"News_Source" json:"-"`
	Body        []string    `xml:"-" json:"body"`
	Image       model.Image `xml:"-" json:"image"`
	Keywords    []string    `xml:"-" json:"keywords"`
}

func (i *Item) String() string {
	b, _ := json.MarshalIndent(i, "", "\t")
	return string(b)
}

func (i *Item) Entry() (string, *Item) {
	return i.URL, i
}

func (i *Item) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"body":        i.Body,
		"description": strings.TrimSpace(i.Description),
		"image":       i.Image,
		"publishedAt": i.PublishedAt,
		"source":      strings.TrimSpace(i.Source),
		"title":       strings.TrimSpace(i.Title),
		"url":         i.URL,
		"keywords":    i.Keywords,
	})
}

func (i *Item) containsKeyword(keyword string) bool {

	keyword = strings.ToLower(keyword)

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

	for _, p := range i.Body {
		if strings.Contains(strings.ToLower(p), keyword) {
			log.Info().
				Str("keyword", keyword).
				Str("body", p).
				Msg("article body contains keyword")
			return true
		}
	}

	log.Info().Str("keyword", keyword).Msg("article does not contain keyword")
	return false
}

func (i *Item) IsBlacklisted(args []string) bool {
	for _, arg := range args {
		if i.containsKeyword(arg) {
			log.Trace().Str("keyword", arg).Msg("blacklisted")
			return true
		}
	}
	return false
}

// ProcessXML improves raw XML data
// - decodes the URL if it's from Google or Bing
// - removes HTML from the description
// - removes the source from the title
func (i *Item) ProcessXML() {

	if strings.Contains(i.URL, "bing.com") {
		i.URL = decodeBingURL(i.URL)
	} else if strings.Contains(i.URL, "google.com") {
		if s, err := decodeGoogleURL(i.URL); err != nil {
			log.Warn().Err(err).Msg("failed to decode google")
		} else {
			i.URL = s
		}
	}

	// populate a potentially empty source
	if i.Source == "" {
		i.Source = i.NewsSource
		if i.Source == "" {
			i.Source = util.Domain(i.URL)
		}
	}

	// remove source from the title if it exists
	i.Title = strings.TrimSuffix(i.Title, " - "+i.Source)

	// remove html from the description
	if idx := strings.Index(i.Description, `</a>`); idx > 0 {
		i.Description = i.Description[:idx]
		i.Description = i.Description[strings.LastIndex(i.Description, ">")+1:]
	}

	// populate a potentially empty publishedAt
	if i.PublishedAt == nil {
		i.PublishedAt = model.Now()
	}
}

func (i *Item) ProcessDoc(doc *model.Document) {

	uniqueParagraphs := make(map[string]int)

	var txt string
	doc.Find("p").Each(func(idx int, s *goquery.Selection) {
		txt = strings.TrimSpace(s.Text())
		if txt == "" ||
			strings.Count(strings.ToLower(txt), strings.ToLower(i.Source)) > 1 ||
			parser.Skip(txt) {
			return
		}
		uniqueParagraphs[txt] = idx
	})

	orderedParagraphs := make(map[int]string)
	for k, v := range uniqueParagraphs {
		orderedParagraphs[v] = k
	}

	keys := slices.Sorted(maps.Keys(orderedParagraphs))
	i.Body = []string{}
	for _, k := range keys {
		i.Body = append(i.Body, orderedParagraphs[k])
	}

	doc.SetMeta()

	i.Image = doc.GetImage()

	if i.Description == "" {
		i.Description, _ = doc.MetaDescription()
	}
	if i.Title == "" {
		i.Title, _ = doc.MetaTitle()
	}
	if i.Source == "" {
		i.Source, _ = doc.MetaSite()
	}
	keywordsStr, _ := doc.MetaKeywords()
	if keywordsStr == "" {
		return
	}
	keywordsArr := strings.Split(keywordsStr, ", ")
	if len(keywordsArr) == 0 {
		keywordsArr = strings.Split(keywordsStr, ",")
	}
	keywordsMap := make(map[string]bool)
	for _, k := range keywordsArr {
		if k = strings.TrimSpace(k); k != "" {
			keywordsMap[k] = true
		}
	}
	i.Keywords = slices.Sorted(maps.Keys(keywordsMap))
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

type Prowler struct {

	// ctx is the context of the browser, which is used to run the browser and the page
	ctx playwright.BrowserContext

	e *entity.News
}

func New(topic string, ctx playwright.BrowserContext) *Prowler {
	return &Prowler{
		ctx: ctx,
		e:   entity.NewNews(topic),
	}
}

func (p *Prowler) Prowl(userID ulid.ULID, workedAt time.Time, blacklist []string) {
	defer func() {
		em.SaveNews(userID, p.e)
	}()
	p.prowl(workedAt, blacklist)
}

func (p *Prowler) prowl(workedAt time.Time, blacklist []string) {
	log.Info().Msgf("processing news worker %s", p.e.Topic)

	q := strings.ReplaceAll(p.e.Topic, ` `, `+`)
	urls := []string{
		fmt.Sprintf("https://www.bing.com/search?format=rss&q=%s", q),
		fmt.Sprintf("https://www.bing.com/news/search?format=rss&q=%s", q),
		fmt.Sprintf("https://news.google.com/rss/search?q=%s&hl=en-US&gl=US&ceid=US:en", q),
	}

	var aa []any

	var rss RSS
	var doc *model.Document
	for _, u := range urls {

		b, err := https.Get(u)
		if err != nil {
			log.Err(err).Str("url", u).Msg("Failed to fetch RSS feed")
			continue
		}

		if strings.Contains(u, "bing.com") {
			b = []byte(bingRegexp.ReplaceAllStringFunc(string(b), func(s string) string {
				return strings.ReplaceAll(s, ":", "_")
			}))
		}

		if err = xml.Unmarshal(b, &rss); err != nil {
			log.Err(err).Str("url", u).Msg("Failed to unmarshal RSS feed")
			continue
		}

		for _, i := range rss.Channel.Items {

			// process the news item before anything else
			i.ProcessXML()

			log.Info().Msgf("working news rss item %s", i.URL)

			// fail fast if this news item could be a duplicate
			if i.PublishedAt.Before(workedAt) {
				log.Debug().Str("url", i.URL).Msg("news item is older than last processed")
				continue
			}

			// get an HTML document for this news item
			if doc, err = pw.Document(p.ctx, i.URL); err != nil {
				log.Warn().Err(err).Str("url", i.URL).Msg("Failed to fetch news document")
				continue
			}

			// process the HTML document and check if it's blacklisted'
			if i.ProcessDoc(doc); i.IsBlacklisted(blacklist) {
				continue
			}

			aa = append(aa, i)
		}
	}

	log.Info().Msgf("processed news worker %s", p.e.Topic)

	em.SaveNews()
}
