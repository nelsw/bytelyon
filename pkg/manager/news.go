package manager

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/nelsw/bytelyon/pkg/client"
	"github.com/nelsw/bytelyon/pkg/db"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/s3"
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

	log.Info().Msgf("processing news worker %s", j.bot.Target)

	q := strings.ReplaceAll(j.bot.Target, ` `, `+`)
	urls := []string{
		fmt.Sprintf("https://www.bing.com/search?format=rss&q=%s", q),
		fmt.Sprintf("https://www.bing.com/news/search?format=rss&q=%s", q),
		fmt.Sprintf("https://news.google.com/rss/search?q=%s&hl=en-US&gl=US&ceid=US:en", q),
	}

	for _, u := range urls {
		j.doNewsFeed(u)
	}

	log.Info().Msgf("processed news worker %s", j.bot.Target)
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
		wg.Go(func() { j.doNewsFeedItem(i) })
	}
	wg.Wait()
}

func (j *Job) doNewsFeedItem(i *model.Item) {

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
	content, screenshot := j.fetchNewsArticle(i.URL)

	// update the result if article content is ok
	if j.handleNewsContent(r, content) ||
		j.handleNewsScreenshot(r, screenshot) ||
		j.handleNewsImage(r) {
		j.updateNewsResult(r)
	}
}

func (j *Job) fetchNewsArticle(s string) (string, []byte) {
	content, screenshot := j.fetchNewsPage(s)
	if content == "" {
		content = j.fetchNewsHTML(s)
	}
	return content, screenshot
}

func (j *Job) fetchNewsPage(s string) (content string, img []byte) {
	page, err := client.NewPage(j.ctx)
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
	)
	if err := db.PutItem(r); err != nil {
		log.Warn().Err(err).Msgf("failed to save news item bot result %s", i)
	}
	return r
}

func (j *Job) handleNewsContent(r *model.BotResult, s string) (ok bool) {

	if s == "" {
		log.Warn().Msg("no news content to process")
		return
	}

	log.Info().Str("url", r.GetStr("url")).Msg("processing item html")

	key, err := s3.PutPrivateBotData(j.bot, fmt.Sprintf("content/%s.html", r.ID), []byte(s))
	if err != nil {
		log.Warn().Err(err).Msg("Failed to save news article html")
		return
	}
	r.Set("content", key)

	var doc *model.Document

	if doc, err = model.NewDocument(r.ID, s); err != nil {
		log.Warn().Err(err).Msg("failed to create doc to hydrate news HTML")
		return
	}

	var body string
	for _, p := range doc.Paragraphs {
		if strings.Count(p, r.GetStr("source")) > 1 ||
			strings.Contains(p, "RELATED:") ||
			strings.Contains(p, "Related:") {
			continue
		}
		body += p + "\n"
	}
	r.Set("body", body)

	if v, k := doc.MetaTitle(); k {
		r.Set("title", v)
	}
	if v, k := doc.MetaDescription(); k {
		r.Set("description", v)
	}
	if v, k := doc.MetaImage(); k {
		r.Set("image", v)
	}
	if v, k := doc.MetaImageAlt(); k {
		r.Set("imageAlt", v)
	}
	if v, k := doc.MetaKeywords(); k {
		r.Set("keywords", v)
	}
	log.Info().Msg("processed news HTML")

	return true
}

func (j *Job) handleNewsScreenshot(r *model.BotResult, b []byte) (ok bool) {

	if len(b) == 0 {
		log.Warn().Msg("no news screenshot to process")
		return
	}

	// define the s3 bucket key for article screenshot
	key, err := s3.PutPublicBotData(j.bot, fmt.Sprintf("screenshot/%s.png", r.ID), b)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to save news article screenshot")
		return
	}

	r.Set("screenshot", key)
	return true
}

func (j *Job) handleNewsImage(r *model.BotResult) (ok bool) {

	src := r.GetStr("image")
	if src == "" {
		log.Warn().Msg("no news image to process")
		return
	} else if !util.IsImageFile(src) {
		log.Warn().Str("image", src).Msg("news image is not a recognized file type")
		return
	}

	out, err := client.Get(src)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to download news article")
		return
	}

	var b []byte
	if b, err = util.ToPng(out); err != nil {
		log.Warn().Err(err).Msg("Failed to convert news article image to png")
		return
	}

	var key string
	if key, err = s3.PutPublicBotData(j.bot, fmt.Sprintf("image/%s.png", r.ID), b); err != nil {
		log.Warn().Err(err).Msg("Failed to save news article image")
		return
	}

	r.Set("image", key)

	return true
}

func (j *Job) updateNewsResult(r *model.BotResult) {
	if err := db.PutItem(r); err != nil {
		log.Warn().Err(err).Msg("Failed to update news item bot result")
	} else {
		log.Info().Msg("updated news item bot result")
	}
}
