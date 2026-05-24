package sitemap

import (
	"strings"
	"sync"
	"time"

	"github.com/nelsw/bytelyon/internal/pw"
	"github.com/nelsw/bytelyon/pkg/id"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/snippet"
	"github.com/nelsw/bytelyon/pkg/url"
	"github.com/oklog/ulid/v2"
	"github.com/playwright-community/playwright-go"
)

type job struct {
	domain string

	// capacitor limits the number of concurrent requests to keep headed browser tabs from crashing.
	capacitor *model.Capacitor

	// BrowserContext is the context of the browser, which is used to scrape the browser and the page.
	ctx playwright.BrowserContext

	// Visited is a map of snippet URLs to snippet IDs to prevent duplicate routines and data.
	visisted *model.SyncMap[string, ulid.ULID]

	// WaitGroup manages the job routines
	wg sync.WaitGroup
}

func (m *Model) Run(ctx playwright.BrowserContext) {

	j := &job{
		domain:    m.Domain,
		ctx:       ctx,
		visisted:  model.NewSyncMap[string, ulid.ULID](),
		capacitor: model.NewCapacitor(25),
	}

	j.wg.Go(func() { j.scrape("https://"+j.domain, 5) })
	j.wg.Wait()

	for k, v := range j.visisted.ToMap() {
		m.Entries[k] = append(m.Entries[k], v)
	}
	m.Save()
}

func (j *job) scrape(u string, d int) {

	// check if we're at the depth limit or if we've already visited this URL
	if d <= 0 || j.visisted.Has(u) {
		return
	}

	pid := id.New()
	j.visisted.Set(u, pid)

	for !j.capacitor.Inc() {
		time.Sleep(500 * time.Millisecond)
	}
	defer j.capacitor.Dec()

	content, screenshot := pw.Scrape(u, j.ctx)

	snip := snippet.New(pid, u, content, screenshot)
	snip.Create()

	for _, a := range j.crawl(u, snip.Links) {
		j.wg.Go(func() { j.scrape(a, d-1) })
	}
}

func (j *job) crawl(u string, links []string) (arr []string) {
	for _, link := range links {

		// if the link is an insecure URL
		if strings.HasPrefix(link, "http://") {
			continue
		}

		// if the link is empty or root
		if link == "" || link == "/" {
			continue
		}

		// if the link is relative to the root urls
		if strings.HasPrefix(link, "/") {
			arr = append(arr, "https://"+j.domain+link)
			continue
		}

		// if the link is a url; check the host equals our domain
		if host := url.Host(link); host != "" && host != j.domain {
			continue
		}

		// if the link is a secure URL
		if strings.HasPrefix(link, "https://"+j.domain) {
			arr = append(arr, link)
			continue
		}

		// if the link is missing URL protocol
		if strings.HasPrefix(link, j.domain) {
			arr = append(arr, "https://"+link)
			continue
		}

		// else the link is relative to this url
		if l, _, ok := strings.Cut(link, "/"); ok {
			arr = append(arr, u+"/"+l+"/"+link)
		} else {
			arr = append(arr, u+"/"+link)
		}
	}

	return
}
