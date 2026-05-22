package sitemap

import (
	"strings"
	"sync"
	"time"

	"github.com/nelsw/bytelyon/pkg/https"
	"github.com/nelsw/bytelyon/pkg/model"
	. "github.com/nelsw/bytelyon/pkg/util"
	"github.com/oklog/ulid/v2"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

// Prowler is a crawler that recursively visits all the URLs on a website using a virtual browser.
type Prowler struct {

	// Sitemap is the prowler model
	*model.Sitemap

	// depth is the maximum number of levels to crawl from the root URL.
	depth int

	// wg manages the state of prowler goroutines
	wg sync.WaitGroup

	// ctx is the context of the browser, which is used to run the browser and the page
	ctx playwright.BrowserContext

	// urls is a map of visited URLs to prevent duplicate visits for this prowl
	urls *model.SyncMap[string, bool]

	// capacitor limits the number of concurrent requests to keep headed browser tabs from crashing
	capacitor *model.Capacitor
}

// New returns a new Prowler instance.
func New(userID ulid.ULID, domain string, ctx playwright.BrowserContext) *Prowler {
	return &Prowler{
		depth:     5, // todo - make configurable
		ctx:       ctx,
		Sitemap:   new(model.Sitemap).From(userID, domain),
		urls:      model.NewSyncMap[string, bool](),
		capacitor: model.NewCapacitor(15),
	}
}

func (p *Prowler) Prowl() {
	p.wg.Go(func() { p.prowl(p.URL(), p.depth) })
	p.wg.Wait()
	log.Info().Msgf("prowled sitemap %s", p.Domain)
}

func (p *Prowler) prowl(url string, depth int) {

	// check if we're at the depth limit or if we've already visited this URL
	if depth <= 0 || p.urls.Has(url) {
		return
	}

	p.urls.Set(url, true)

	for !p.capacitor.Inc() {
		time.Sleep(time.Second)
	}
	defer p.capacitor.Dec()

	page := new(model.Page).Scrape(url, p.ctx)
	if page == nil {
		return
	}
	page.Save()

	p.Add(page)
	p.Save()

	var urls []string
	for _, link := range page.Links {

		// if the link is empty or root
		if link == "" || link == "/" {
			continue
		}

		// if the link is relative to the root url
		if strings.HasPrefix(link, "/") {
			urls = append(urls, p.URL()+link)
			continue
		}

		// if the link is a url; check the host equals our domain
		if host := Host(link); host != "" && host != p.Domain {
			continue
		}

		// if the link is a secure URL
		if strings.HasPrefix(link, p.URL()) {
			urls = append(urls, link)
			continue
		}

		// if the link is missing URL protocol
		if strings.HasPrefix(link, p.Domain) {
			urls = append(urls, https.Protcol(link))
			continue
		}

		// else the link is relative to this url
		if l, _, ok := strings.Cut(link, "/"); ok {
			urls = append(urls, url+"/"+l+"/"+link)
		} else {
			urls = append(urls, url+"/"+link)
		}
	}

	for _, u := range urls {
		p.wg.Go(func() { p.prowl(u, p.depth-1) })
	}
}
