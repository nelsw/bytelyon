package sitemaps

import (
	"regexp"
	"strings"
	"sync"

	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/repo"
	"github.com/nelsw/bytelyon/pkg/service/documents"
	"github.com/nelsw/bytelyon/pkg/service/pages"
	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

var (
	badAnchorRegex = regexp.MustCompile(`^(#|mailto:|tel:).*`)
	badExtRegex    = regexp.MustCompile(`^.*\.(jpeg|png|gif|jpg|pdf)$`)
)

func Create(domain string, depth int, ctx playwright.BrowserContext) (err error) {

	m := New(domain, depth)

	if err = repo.SaveSitemap(m); err != nil {
		return
	}

	for _, url := range m.URLs.Slice() {
		if err = pages.Create(url, ctx); err != nil {
			log.Err(err).Msg("failed to create sitemap page")
		}
	}
	return
}

func New(domain string, depth int) *model.Sitemap {

	var ƒ func(m *model.Sitemap, wg *sync.WaitGroup, u string, d int)

	ƒ = func(m *model.Sitemap, wg *sync.WaitGroup, u string, d int) {

		if d <= 0 || m.URLs.Has(u, true) {
			return
		}
		m.URLs.Add(u, true)

		doc, err := documents.New(u)
		if err != nil {
			return
		}

		for _, href := range doc.GetHREFs() {

			href = strings.TrimSpace(href)
			href = strings.TrimSuffix(href, "/")

			if href == "" ||
				badExtRegex.MatchString(href) ||
				badAnchorRegex.MatchString(href) ||
				strings.HasSuffix(href, "@"+m.Domain) {
				continue
			}

			if util.Domain(href) == m.Domain {
				wg.Go(func() { ƒ(m, wg, href, d-1) })
			}

			if strings.HasPrefix(href, "?") || strings.HasPrefix(href, "/") {
				wg.Go(func() { ƒ(m, wg, "https://"+m.Domain+href, d-1) })
			}
		}
	}

	m := model.NewSitemap(domain)

	var wg sync.WaitGroup
	wg.Go(func() { ƒ(m, &wg, "https://"+domain, depth) })
	wg.Wait()

	return m
}
