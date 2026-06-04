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

	log.Err(Save(userID, domain, m.Clone())).Msgf("worked sitemap %s", domain)
}

func work(
	capacitor *model.Counter,
	ctx playwright.BrowserContext,
	smap *model.SyncMap[string, bool],
	wg *sync.WaitGroup,
	domain string,
	url string,
	depth int,
) {

	// check if we're at the depth limit or if we've already visited this URL
	if depth <= 0 || smap.Has(url) {
		return
	}

	smap.Put(url, true)

	for !capacitor.Inc() {
		time.Sleep(500 * time.Millisecond)
	}

	content, screenshot := pw.Scrape(url, ctx)
	capacitor.Dec()

	if content == "" {
		smap.Drop(url)
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
			wg.Go(func() { work(capacitor, ctx, smap, wg, domain, u, depth-1) })
		}
	}
}

func pageLink(domain, href string) (string, bool) {

	// trim whitespace, lowercase, and remove trailing slash
	href = urls.Clean(href)

	// if the href is ...
	if href == "" || // root
		path.Ext(href) != "" || // file
		urls.IsBrowserFunction(href) || // browser function
		strings.HasPrefix(href, "#") || // fragment
		strings.HasPrefix(href, "http://") || // insecure
		(strings.HasPrefix(href, "https://") && !strings.HasPrefix(href, "https://"+domain)) { // outbound
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
