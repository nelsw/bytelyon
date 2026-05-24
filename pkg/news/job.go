package news

import (
	"maps"
	"slices"
	"sync"
	"time"

	"github.com/nelsw/bytelyon/internal/pw"
	"github.com/nelsw/bytelyon/pkg/article"
	"github.com/nelsw/bytelyon/pkg/document"
	"github.com/nelsw/bytelyon/pkg/page"
	"github.com/oklog/ulid/v2"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

func (m *Model) Run(ctx playwright.BrowserContext, exclude map[string]bool) {

	var after time.Time
	if ids := slices.SortedFunc(maps.Values(m.Entries), func(a, z ulid.ULID) int { return a.Compare(z) }); len(ids) > 0 {
		after = ids[0].Timestamp()
	}

	var arr []*article.Model
	for _, a := range article.FromRSS(m.Topic, after, exclude) {
		if _, ok := m.Entries[a.URL]; !ok {
			m.Entries[a.URL] = a.ID
			arr = append(arr, a)
		}
	}

	var wg sync.WaitGroup
	for _, a := range arr {
		wg.Go(func() {

			content, screenshot := pw.Scrape(a.URL, ctx)
			if len(content) == 0 {
				return
			}

			doc := document.New(content)
			if a.Description == "" {
				a.Description = doc.Description()
			}
			if len(a.Image) == 0 {
				a.Image = doc.Image()
			}
			if a.Image.GetSrc() == "" {
				a.Image.SetSrc(doc.ImageSrc())
			}
			if a.Image.GetAlt() == "" {
				a.Image.SetAlt(doc.ImageAlt())
			}
			a.Keywords = doc.Keywords()
			a.Meta = doc.Meta
			a.Body = doc.Paragraphs

			if err := page.Create(a.URL, a.ID, content, screenshot, a); err != nil {
				log.Warn().Err(err).Msg("failed to create news")
				return
			}
		})
	}
	wg.Wait()

	m.Save()
}
