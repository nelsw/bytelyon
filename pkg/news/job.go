package news

import (
	"sync"
	"time"

	"github.com/nelsw/bytelyon/pkg/article"
	"github.com/playwright-community/playwright-go"
)

func (m *Model) Run(ctx playwright.BrowserContext, after time.Time, exclude map[string]bool) {

	defer m.Save()

	var arr article.Models
	for _, a := range article.FromRSS(m.Topic, after, exclude) {
		if _, ok := m.Entries[a.URL]; !ok {
			m.Entries[a.URL] = a.ID
			arr = append(arr, a)
		}
	}

	var wg sync.WaitGroup
	for _, a := range arr {
		wg.Go(func() {
			a.Fill(ctx)
			a.Create()
		})
	}
	wg.Wait()
}
