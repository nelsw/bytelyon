package sitemap

import (
	"path"
	"strings"
	"sync"
	"time"

	"github.com/nelsw/bytelyon/pkg/document"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/page"
	"github.com/nelsw/bytelyon/pkg/pw"
	"github.com/nelsw/bytelyon/pkg/snippet"
	"github.com/nelsw/bytelyon/pkg/util/urls"
	"github.com/oklog/ulid/v2"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

func Work(ctx playwright.BrowserContext, userID ulid.ULID, domain string) {

	m := model.NewSyncMap[string, bool]()

	var wg sync.WaitGroup

	wg.Go(func() {
		work(
			model.NewCounter(25),
			ctx,
			m,
			&wg,
			domain,
			"https://"+domain,
			5,
		)
	})
	wg.Wait()

	if err := Save(userID, domain, m.Clone()); err != nil {
		log.Warn().Err(err).Msg("failed to save sitemap")
	}
}

func work(
	capacitor *model.Counter,
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

	urls.Put(url, true)

	for !capacitor.Inc() {
		time.Sleep(500 * time.Millisecond)
	}
	defer capacitor.Dec()

	content, screenshot := pw.Scrape(url, ctx)
	if content == "" {
		return
	}

	doc := document.New(content)

	snip := snippet.New(url, doc.Title(), doc.Meta)
	if err := page.SaveObject(snip.URL, snip.ID, snip); err != nil {
		return
	} else if err = page.SaveScreenshot(snip.URL, snip.ID, screenshot); err != nil {
		return
	}

	for _, href := range doc.HREFs() {
		if u, ok := pageLink(domain, href); ok {
			wg.Go(func() { work(capacitor, ctx, urls, wg, domain, u, depth-1) })
		}
	}
}

func pageLink(domain, href string) (string, bool) {

	// trim whitespace, lowercase, and remove trailing slash
	href = urls.Clean(href)

	// if the href is ...
	if path.Ext(href) != "" || // file
		href == "" || href == "/" || // root
		urls.Domain(href) != domain || // outbound
		urls.IsBrowserFunction(href) || // browser function
		strings.HasPrefix(href, "#") || // fragment
		strings.HasPrefix(href, "http://") { // insecure
		return "", false
	}

	// if the link is a secure URL
	if strings.HasPrefix(href, "https://"+domain) {
		return href, true
	}

	// if the link is missing URL protocol
	if strings.HasPrefix(href, domain) {
		return "https://" + href, true
	}

	// if the link is relative to the root urls
	if strings.HasPrefix(href, "/") {
		return "https://" + domain + href, true
	}

	// else the link is relative to this url
	return "https://" + domain + "/" + href, true
}
