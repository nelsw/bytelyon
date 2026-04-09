package manager

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/joho/godotenv"
	"github.com/nelsw/bytelyon/pkg/client"
	"github.com/nelsw/bytelyon/pkg/logs"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/pw"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func TestSiteMapper(t *testing.T) {

	target := "https://li-fire.com/"

	m := NewMapper(&fetcher{}, target)
	m.Add()
	m.Map(target, 3)
	m.Wait()

	urls := m.Relative()

	log.Info().
		Str("target", target).
		Msgf("URLs found [%d]", len(urls))

	//result := j.bot.NewBotResult("relative", urls)

	//if err := db.PutItem(result); err != nil {
	//	log.Err(err).Msg("failed to put sitemap result, pressing on ...")
	//} else {
	//	log.Info().Msg("sitemap result put")
	//}
	godotenv.Load()
	logs.Init()
	pw.Init()
	headless, headed, err := pw.NewBrowsers()
	if err != nil {
		panic(err)
	}

	defer func() {
		headless.Close()
		headed.Close()
		pw.Client.Stop()
	}()

	var ctx playwright.BrowserContext
	ctx, err = client.NewContext(headed, nil)
	assert.NoError(t, err)
	handlePages(urls, ctx)
}

func handlePages(urls []string, ctx playwright.BrowserContext) {

	var pages []*model.Page

	defer func() {

		log.Debug().Msgf("sitemap pages to save [%d]", len(pages))
		if len(pages) == 0 {
			return
		}

		//r.Data["pages"] = pages
		//
		//if err := db.PutItem(r); err != nil {
		//	log.Warn().Err(err).Msg("failed to save sitemap pages")
		//} else {
		//	log.Info().Msgf("sitemap pages saved [%d]", len(pages))
		//}
	}()

	var key string
	for idx, url := range urls {

		title, _, _, err := pw.FetchPageData(url, ctx)
		if err != nil {
			log.Warn().Err(err).Str("url", url).Msg("failed to handle page")
			continue
		}

		p := &model.Page{
			Title: title,
			URL:   url,
		}

		key = fmt.Sprintf("content/%s/%d.html", "test-id", idx)
		fmt.Println(key)
		//if p.HTML, err = s3.PutPublicBotResultData(r, key, []byte(src)); err != nil {
		//	log.Warn().Err(err).Str("url", url).Msg("failed to save page html")
		//}
		//
		key = fmt.Sprintf("image/%s/%d.png", "test-id", idx)
		fmt.Println(key)
		//if p.IMG, err = s3.PutPublicBotResultData(r, key, img); err != nil {
		//	log.Warn().Err(err).Str("url", url).Msg("failed to save page image")
		//}

		log.Info().Str("url", url).Msg("page handled")
		pages = append(pages, p)
	}

	b, err := json.MarshalIndent(pages, "", "  ")
	fmt.Println(
		string(b),
		err,
	)
}
