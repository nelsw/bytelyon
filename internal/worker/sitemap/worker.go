package sitemap

import (
	"github.com/nelsw/bytelyon/internal/model"
	"github.com/nelsw/bytelyon/internal/util"
)

type Worker struct {
	url string
}

func New(url string) Worker {
	return Worker{url}
}

func (w Worker) Work() *model.Sitemap {

	m := NewMapper(&fetcher{}, w.url)
	m.Add()
	m.Map(w.url, 3)
	m.Wait()

	return &model.Sitemap{
		URL:      w.url,
		Domain:   util.Domain(w.url),
		Relative: m.Relative(),
		Remote:   m.Remote(),
	}
}
