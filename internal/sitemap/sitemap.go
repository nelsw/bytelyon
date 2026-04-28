package sitemap

import (
	"errors"
	"regexp"
	"strings"

	"github.com/nelsw/bytelyon/pkg/db"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/service/documents"
	"github.com/nelsw/bytelyon/pkg/service/pages"
	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

var (
	badHref = regexp.MustCompile(`^(#|mailto:|tel:).*`)
	badExt  = regexp.MustCompile(`^.*\.(jpeg|png|gif|jpg|pdf)$`)
)

type Job struct {
	Domain string
	ctx    playwright.BrowserContext
}

func New(s string, ctx playwright.BrowserContext) *Job {
	return &Job{s, ctx}
}

func (j *Job) Work() {
	m := model.NewSitemap(j.Domain)
	crawl(m, "https://"+j.Domain, 5)
	save(m, j.ctx)
}

func crawl(m *model.Sitemap, u string, d int) {
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
			badExt.MatchString(href) ||
			badHref.MatchString(href) ||
			strings.HasSuffix(href, "@"+m.Domain) {
			continue
		}

		if util.Domain(href) == m.Domain {
			crawl(m, href, d-1)
		}

		if strings.HasPrefix(href, "?") || strings.HasPrefix(href, "/") {
			crawl(m, "https://"+m.Domain+href, d-1)
		}
	}
}

func save(newM *model.Sitemap, ctx playwright.BrowserContext) (errs error) {

	newUrls := newM.URLs.Slice()

	oldM, _ := db.Get(newM)
	oldM.URLs.Add(newUrls)

	if err := db.Put(oldM); err != nil {
		log.Err(err).Msg("failed to save sitemap")
		errs = errors.Join(errs, err)
	}

	for _, url := range newUrls {
		if err := pages.Create(url, ctx); err != nil {
			log.Err(err).Msg("failed to create sitemap page")
			errs = errors.Join(errs, err)
		}
	}

	return
}
