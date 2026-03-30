package job

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
	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

var (
	bingRegexp = regexp.MustCompile("</?News(:\\w+)>")
)

type RSS struct {
	Channel struct {
		Items []*model.Article `xml:"item"`
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
		wg.Go(func() { j.doNewsFeedArticle(ctx, i) })
	}
	wg.Wait()
}

func (j *Job) doNewsFeedArticle(ctx playwright.BrowserContext, article *model.Article) {

	log.Info().Object("article", article).Msg("Processing RSS item (news article)")

	// fail fast if we've seen this article
	if article.IsOldNews(j.bot.WorkedAt) {
		return
	}

	// work some magic to circumvent protected urls
	article.DecodeURL()

	if err := j.FetchDynamicHTML(ctx, article); err != nil {
		j.FetchStaticHTML(article)
	}

	// include HTML data
	article.ProcessHTML()

	// scrub all the data
	article.ScrubDetails()

	// now that the item is populated with data,
	// check article for blacklisted keywords
	if article.IsBlacklisted(j.bot.BlackList) {
		return
	}

	// instantiate a new bot result
	result := j.bot.NewBotResult(
		"url", article.URL,
		"title", article.Title,
		"source", article.Source,
		"description", article.Description,
		"publishedAt", article.Time.String(),
		"body", article.Body,
	)

	// save article HTML if it exits and define the path on the result
	if article.Content != "" {
		// define the s3 bucket key for article html
		key := fmt.Sprintf("users/%s/bots/news/%s/content/%s.html",
			j.bot.UserID,
			j.bot.Target,
			result.ID,
		)

		if err := client.PutObject(j.ctx, j.s3, "bytelyon-public", key, []byte(article.Content)); err != nil {
			log.Warn().Err(err).Object("article", article).Msg("Failed to save news article html")
			article.Content = ""
		} else {
			result.Data["content"] = key
			article.Content = key
		}
	}

	// save article image if it exists and define the path on the result
	if article.Image != "" {

		res, err := http.Get(article.Image)
		if err != nil {
			log.Warn().Err(err).Object("article", article).Msg("Failed to download news article")
		} else {
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(res.Body)

			var b []byte
			if b, err = io.ReadAll(res.Body); err != nil {
				log.Warn().Err(err).Object("article", article).Msg("Failed to read news article")
			} else {
				// define the s3 bucket key for article image
				key := fmt.Sprintf("users/%s/bots/news/%s/image/%s%s",
					j.bot.UserID,
					j.bot.Target,
					result.ID,
					util.Extension(article.Image),
				)

				if err = client.PutObject(j.ctx, j.s3, "bytelyon-public", key, b); err != nil {
					log.Warn().Err(err).Object("article", article).Msg("Failed to save news article image")
				} else {
					result.Data["image"] = key
					article.Image = key
				}
			}
		}
	}

	if article.Screenshot != "" {
		// define the s3 bucket key for article html
		key := fmt.Sprintf("users/%s/bots/news/%s/screenshot/%s.png",
			j.bot.UserID,
			j.bot.Target,
			result.ID,
		)

		if err := client.PutObject(j.ctx, j.s3, "bytelyon-public", key, []byte(article.Screenshot)); err != nil {
			log.Warn().Err(err).Object("article", article).Msg("Failed to save news article screenshot")
			article.Screenshot = ""
		} else {
			result.Data["screenshot"] = key
			article.Screenshot = key
		}
	}

	// save the result
	if err := db.PutItem(result); err != nil {
		log.Warn().Err(err).Object("article", article).Msg("Failed to save news article")
	} else {
		log.Info().Object("article", article).Msg("News article saved")
	}
}

func (j *Job) FetchStaticHTML(a *model.Article) {
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
		Msg("got static html for news")
}

func (j *Job) FetchDynamicHTML(ctx playwright.BrowserContext, a *model.Article) error {
	page, err := client.NewPage(ctx)
	if err != nil {
		log.Err(err).Msg("failed to create new page for news article with pw")
		return err
	}
	defer func(page playwright.Page) {
		_ = page.Close()
	}(page)

	var resp playwright.Response
	if resp, err = client.GoTo(page, a.URL); err != nil {
		log.Err(err).Str("url", a.URL).Msg("failed to go to news article url with pw")
		return err
	}

	if client.IsRequestBlocked(resp) || client.IsPageBlocked(page) {
		log.Warn().Str("url", a.URL).Msg("page/request is blocked for news article with pw")
		return err
	}

	log.Info().
		Str("url", a.URL).
		Msg("got dynamic html for news")

	if a.Content, err = page.Content(); err != nil {
		log.Warn().Err(err).Msg("Failed to get news article Page Content")
		return err
	}

	var screenshot []byte
	if screenshot, err = page.Screenshot(playwright.PageScreenshotOptions{FullPage: util.Ptr(true)}); err != nil {
		log.Warn().Err(err).Msg("Failed to get news article Screenshot")
	} else {
		a.Screenshot = string(screenshot)
	}
	_ = page.Close()

	return nil
}
