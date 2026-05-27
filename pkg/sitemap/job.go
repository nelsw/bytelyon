package sitemap

import (
	"sync"
	"time"

	"github.com/nelsw/bytelyon/internal/pw"
	"github.com/nelsw/bytelyon/pkg/id"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/snippet"
	"github.com/oklog/ulid/v2"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

func Work(ctx playwright.BrowserContext, userID ulid.ULID, domain string) {

	urls := model.NewSyncMap[string, bool]()

	var wg sync.WaitGroup

	wg.Go(func() {
		work(
			model.NewCapacitor(25),
			ctx,
			urls,
			&wg,
			domain,
			"https://"+domain,
			5,
		)
	})
	wg.Wait()

	if err := Save(userID, domain, urls); err != nil {
		log.Warn().Err(err).Msg("failed to save sitemap")
	}
}

func work(
	capacitor *model.Capacitor,
	ctx playwright.BrowserContext,
	urls *model.SyncMap[string, bool],
	wg *sync.WaitGroup,
	domain string,
	url string,
	depth int,
) {

	// check if we're at the depth limit or if we've already visited this URL
	if depth <= 0 || urls.Has(url) {
		return
	}

	urls.Set(url, false)

	for !capacitor.Inc() {
		time.Sleep(500 * time.Millisecond)
	}
	defer capacitor.Dec()

	content, screenshot := pw.Scrape(url, ctx)
	if content == "" {
		return
	}

	snip := snippet.New(id.New(), url, content, screenshot)
	if err := snip.Save(); err != nil {
		return
	}
	urls.Set(url, true)

	for _, u := range snip.URLs() {
		wg.Go(func() { work(capacitor, ctx, urls, wg, domain, u, depth-1) })
	}
}
