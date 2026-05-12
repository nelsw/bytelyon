package sitemap

import (
	"strings"
	"sync"

	"github.com/nelsw/bytelyon/internal/pw"
	"github.com/nelsw/bytelyon/pkg/em"
	"github.com/nelsw/bytelyon/pkg/entity"
	"github.com/nelsw/bytelyon/pkg/model"
	. "github.com/nelsw/bytelyon/pkg/util"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

// Prowler is a crawler that recursively visits all the URLs on a website using a virtual browser.
type Prowler struct {

	// depth is the maximum number of levels to crawl from the root URL.
	depth int

	// wg manages the state of prowler goroutines
	wg sync.WaitGroup

	// ctx is the context of the browser, which is used to run the browser and the page
	ctx playwright.BrowserContext

	*entity.Sitemap
}

func New(bot *model.Bot, ctx playwright.BrowserContext) *Prowler {
	return &Prowler{
		depth:   5, // todo - make configurable
		ctx:     ctx,
		Sitemap: entity.NewSitemap(bot),
	}
}

func (p *Prowler) Prowl() {
	defer func() {
		em.PutSitemap(p.Sitemap)
	}()
	p.wg.Go(func() { p.prowl("https://"+p.Target, p.depth) })
	p.wg.Wait()
}

func (p *Prowler) prowl(url string, depth int) {

	// check if we're at the depth limit or if we've already visited this URL
	if depth--; depth < 0 || p.Has(url) {
		return
	}

	var urls []string
	for _, link := range p.put(url) {

		// if the link is empty or root
		if link == "" || link == "/" {
			continue
		}

		// if the link is relative to the root url
		if strings.HasPrefix(link, "/") {
			urls = append(urls, "https://"+p.Target+link)
			continue
		}

		// if the link is a url; check the host equals our domain
		if host := Host(link); host != "" && host != p.Target {
			continue
		}

		// if the link is a secure URL
		if strings.HasPrefix(link, "https://"+p.Target) {
			urls = append(urls, link)
			continue
		}

		// if the link is missing URL protocol
		if strings.HasPrefix(link, p.Target) {
			urls = append(urls, "https://"+link)
			continue
		}

		// else the link is relative to this url
		if l, _, ok := strings.Cut(link, "/"); !ok {
			urls = append(urls, url+"/"+link)
		} else {
			urls = append(urls, url+"/"+l+"/"+link)
		}
	}

	for _, u := range urls {
		p.wg.Go(func() { p.prowl(u, depth) })
	}
}

func (p *Prowler) put(url string) (links []string) {

	l := log.With().
		Str("ƒ", "Prowler.put").
		Str("url", url).
		Logger()

	l.Trace().Send()

	page, err := pw.NewPage(p.ctx)
	if err != nil {
		l.Warn().Msgf("NewPage failed: %s", err.Error())
		return
	}

	defer page.Close()

	if err = pw.Visit(page, url); err != nil {
		l.Warn().Msgf("Visit failed: %s", err.Error())
		return
	}

	l.Debug().Msgf("put %s", url)

	return p.Set(url, entity.NewPage(page)).Links
}
