package job

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/nelsw/bytelyon/pkg/client"
	"github.com/nelsw/bytelyon/pkg/db"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

var (
	bingRegexp = regexp.MustCompile("</?News(:\\w+)>")
)

type RSS struct {
	Channel struct {
		Items []*model.Item `xml:"item"`
	} `xml:"channel"`
}

func (j *Job) doNews() {

	log.Info().Msgf("processing news job %s", j.bot.Target)

	q := strings.ReplaceAll(j.bot.Target, ` `, `+`)
	urls := []string{
		fmt.Sprintf("https://www.bing.com/search?format=rss&q=%s", q),
		fmt.Sprintf("https://www.bing.com/news/search?format=rss&q=%s", q),
		fmt.Sprintf("https://news.google.com/rss/search?q=%s&hl=en-US&gl=US&ceid=US:en", q),
	}

	var err error

	var ply *playwright.Playwright
	if ply, err = client.NewPlaywright(); err != nil {
		return
	}
	defer func(ply *playwright.Playwright) {
		_ = ply.Stop()
	}(ply)

	var bro playwright.Browser
	if bro, err = client.NewBrowser(ply, true); err != nil {
		return
	}
	defer func(bro playwright.Browser) {
		_ = bro.Close()
	}(bro)

	var ctx playwright.BrowserContext
	if ctx, err = client.NewContext(bro); err != nil {
		return
	}
	defer func(ctx playwright.BrowserContext) {
		_ = ctx.Close()
	}(ctx)

	for _, u := range urls {
		j.doNewsFeed(ctx, u)
	}

	log.Info().Msgf("processed news job %s", j.bot.Target)
}

func (j *Job) doNewsFeed(ctx playwright.BrowserContext, u string) {
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
		wg.Go(func() { j.doNewsFeedItem(ctx, i) })
	}
	wg.Wait()
}

func (j *Job) doNewsFeedItem(ctx playwright.BrowserContext, i *model.Item) {

	log.Info().Msgf("Processing RSS item %s", i)

	// fail fast if we've seen this article
	if i.IsOldNews(j.bot.WorkedAt) {
		return
	}

	// process xml item and check if it's blacklisted
	if i.ProcessXML(); i.IsBlacklisted(j.bot.BlackList) {
		return
	}

	// create a result from initial data
	r := j.createNewsResult(i)

	// try to fetch the article content
	content, screenshot := j.fetchNewsArticle(ctx, i.URL)

	// update the result if article content is ok
	if j.handleNewsContent(i, r, content) ||
		j.handleNewsScreenshot(r, screenshot) ||
		j.handleNewsImage(i, r) {
		j.updateNewsResult(r)
	}
}

func (j *Job) fetchNewsArticle(ctx playwright.BrowserContext, s string) (string, []byte) {
	content, screenshot := j.fetchNewsPage(ctx, s)
	if content == "" {
		content = j.fetchNewsHTML(s)
	}
	return content, screenshot
}

func (j *Job) fetchNewsPage(ctx playwright.BrowserContext, s string) (content string, img []byte) {
	page, err := client.NewPage(ctx)
	if err != nil {
		log.Err(err).Msg("failed to create new page for news article with pw")
		return
	}
	defer func(page playwright.Page) {
		_ = page.Close()
	}(page)

	var resp playwright.Response
	if resp, err = client.GoTo(page, s); err != nil {
		log.Err(err).Str("url", s).Msg("failed to go to news article url with pw")
		return
	} else if client.IsRequestBlocked(resp) || client.IsPageBlocked(page) {
		log.Warn().Str("url", s).Msg("page/request is blocked for news article with pw")
		return
	}

	log.Info().
		Str("url", s).
		Msg("got dynamic html for news")

	if content, err = page.Content(); err != nil {
		log.Warn().Err(err).Msg("Failed to get news article Page Content")
	}

	if img, err = page.Screenshot(playwright.PageScreenshotOptions{FullPage: util.Ptr(true)}); err != nil {
		log.Warn().Err(err).Msg("Failed to get news article Screenshot")
	}

	return
}

func (j *Job) fetchNewsHTML(s string) string {
	res, err := http.Get(s)
	if err != nil {
		log.Warn().Err(err).Msg("failed to fetch URL to hydrate news HTML")
		return ""
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	var b []byte
	if b, err = io.ReadAll(res.Body); err != nil {
		log.Warn().Err(err).Msg("failed to read news HTML")
		return ""
	}

	log.Info().
		Int("status", res.StatusCode).
		Str("url", s).
		Msg("got static html for news")

	return string(b)
}

func (j *Job) createNewsResult(i *model.Item) *model.BotResult {
	r := j.bot.NewBotResult(
		"url", i.URL,
		"title", i.Title,
		"source", i.Source,
		"description", i.Description,
		"publishedAt", i.Time.String(),
		"body", i.Body,
	)
	if err := db.PutItem(r); err != nil {
		log.Warn().Err(err).Msgf("failed to save news item bot result %s", i)
	}
	return r
}

func (j *Job) handleNewsContent(i *model.Item, r *model.BotResult, s string) (ok bool) {

	if s == "" {
		log.Warn().Msg("no news content to process")
		return
	}

	log.Info().Str("url", i.URL).Msg("processing item html")

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(s))
	if err != nil {
		log.Warn().Err(err).Msg("failed to create doc to hydrate news HTML")
		return
	}

	doc.Find("p").Each(func(idx int, s *goquery.Selection) { i.Body += s.Text() + "\n" })

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
				if v, k = m["content"]; k && util.IsImageFile(v) {
					i.Image = v
				}
			}
		}
		if i.Image == "" {
			if v, k := m["property"]; k && v == "og:image" {
				if v, k = m["content"]; k && util.IsImageFile(v) {
					i.Image = v
				}
			}
		}
	}

	if i.Title == "" {
		doc.Find("title").Each(func(idx int, s *goquery.Selection) { i.Title = s.Text() })
	}

	log.Info().Msg("processed news HTML")

	// define the s3 bucket key for article HTML
	key := fmt.Sprintf("users/%s/bots/news/%s/content/%s.html",
		j.bot.UserID,
		j.bot.Target,
		r.ID,
	)

	if err = client.PutObject(j.ctx, j.s3, "bytelyon-public", key, []byte(s)); err != nil {
		log.Warn().Err(err).Msg("Failed to save news article html")
		return
	}

	r.Data["content"] = key
	return true
}

func (j *Job) handleNewsScreenshot(r *model.BotResult, b []byte) (ok bool) {

	if len(b) == 0 {
		log.Warn().Msg("no news screenshot to process")
		return
	}

	// define the s3 bucket key for article screenshot
	key := fmt.Sprintf("users/%s/bots/news/%s/screenshot/%s.png",
		j.bot.UserID,
		j.bot.Target,
		r.ID,
	)

	if err := client.PutObject(j.ctx, j.s3, "bytelyon-public", key, b); err != nil {
		log.Warn().Err(err).Msg("Failed to save news article screenshot")
		return
	}

	r.Data["screenshot"] = key
	return true
}

func (j *Job) handleNewsImage(i *model.Item, r *model.BotResult) (ok bool) {

	if i.Image == "" {
		log.Warn().Msg("no news image to process")
		return
	}

	res, err := http.Get(i.Image)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to download news article")
		return
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	var b []byte
	if b, err = io.ReadAll(res.Body); err != nil {
		log.Warn().Err(err).Msg("Failed to read news article")
		return
	}

	// define the s3 bucket key for article image
	key := fmt.Sprintf("users/%s/bots/news/%s/image/%s%s",
		j.bot.UserID,
		j.bot.Target,
		r.ID,
		util.Extension(i.Image),
	)

	if err = client.PutObject(j.ctx, j.s3, "bytelyon-public", key, b); err != nil {
		log.Warn().Err(err).Msg("Failed to save news article image")
		return
	}

	r.Data["image"] = key
	i.Image = key

	return true
}

func (j *Job) updateNewsResult(r *model.BotResult) {
	if err := db.PutItem(r); err != nil {
		log.Warn().Err(err).Msg("Failed to update news item bot result")
	} else {
		log.Info().Msg("updated news item bot result")
	}
}
