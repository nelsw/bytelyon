package manager

import (
	"maps"
	"net/http"
	"regexp"
	"slices"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/nelsw/bytelyon/pkg/db"
	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/rs/zerolog/log"
)

var (
	badAnchorRegex = regexp.MustCompile(`^(#|mailto:|tel:).*`)
	badExtRegex    = regexp.MustCompile(`^.*\.(jpeg|png|gif|jpg|pdf)$`)
)

type Crawler interface {
	Crawl(string) []string
}

type fetcher struct{}

func (f *fetcher) Crawl(s string) []string {

	r, err := http.Get(s)
	if err != nil {
		log.Err(err).Send()
		return nil
	}
	defer r.Body.Close()

	var doc *goquery.Document
	if doc, err = goquery.NewDocumentFromReader(r.Body); err != nil {
		log.Err(err).Send()
		return nil
	}

	var ss []string
	doc.Find("a").Each(func(i int, sel *goquery.Selection) {
		ss = append(ss, sel.AttrOr("href", ""))
	})

	return ss
}

type Mapper struct {
	Crawler
	relative map[string]bool
	remote   map[string]bool
	mu       sync.Mutex
	wg       sync.WaitGroup
	url      string
	domain   string
}

func NewMapper(c Crawler, url string) *Mapper {
	return &Mapper{
		Crawler:  c,
		relative: make(map[string]bool),
		remote:   make(map[string]bool),
		url:      url,
		domain:   util.Domain(url),
	}
}

func (m *Mapper) Map(url string, depth int) {

	defer m.wg.Done()

	if depth <= 0 || m.putRelative(url) {
		return
	}

	ss := m.Crawl(url)

	rel, rem := m.Categorize(ss)

	m.putAllRemote(rem)

	for _, u := range rel {
		m.Add()
		go m.Map(u, depth-1)
	}
}

func (m *Mapper) Add() {
	m.wg.Add(1)
}

func (m *Mapper) Wait() {
	m.wg.Wait()
}

func (m *Mapper) putRelative(s string) (ok bool) {

	// ignore self
	if s == m.url || s == m.domain {
		return true
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok = m.relative[s]; !ok {
		m.relative[s] = true
	}
	return ok
}

func (m *Mapper) putAllRemote(urls []string) {
	// todo - blacklist
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, url := range urls {
		m.remote[url] = true
	}
}

func (m *Mapper) Relative() []string {
	return slices.Sorted(maps.Keys(m.relative))
}

func (m *Mapper) Remote() []string {
	return slices.Sorted(maps.Keys(m.remote))
}

func (m *Mapper) Categorize(ss []string) ([]string, []string) {
	var rel, rem []string
	for _, s := range ss {
		s = strings.TrimSpace(s)
		s = strings.TrimSuffix(s, "/")
		if s == "" {
			continue
		}

		if badAnchorRegex.MatchString(s) ||
			badExtRegex.MatchString(s) ||
			strings.HasSuffix(s, "@"+m.domain) {
			continue
		}

		if u := util.Domain(s); u == m.domain {
			rel = append(rel, s)
			continue
		}

		if strings.HasPrefix(s, "?") || strings.HasPrefix(s, "/") {
			rel = append(rel, m.url+s)
			continue
		}

		rem = append(rem, s)
	}
	return rel, rem
}

func (j *Job) doSitemap() {
	m := NewMapper(&fetcher{}, j.bot.Target)
	m.Add()
	m.Map(j.bot.Target, 3)
	m.Wait()

	err := db.PutItem(j.bot.NewBotResult(
		"relative", m.Relative(),
		"remote", m.Remote(),
	))

	log.Err(err).Msg("put sitemap result")
}
