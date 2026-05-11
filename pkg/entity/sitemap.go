package entity

import (
	"github.com/nelsw/bytelyon/pkg/dto"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/oklog/ulid/v2"
)

type Sitemap struct {
	ID     ulid.ULID   `json:"id"`
	Domain string      `json:"domain"`
	Nodes  []*dto.Node `json:"nodes"`
	URLs   []string    `json:"urls"`

	Pages *model.SyncMap[string, *Page] `json:"-"`
}

func NewSitemap(domain string) *Sitemap {
	return &Sitemap{
		ID:     model.NewULID(),
		Domain: domain,
	}
}

func (s *Sitemap) AddPage(p *Page) {
	s.URLs = append(s.URLs, p.URL)
	s.Pages.Set(p.URL, p)
}

func (s *Sitemap) AddURLs(urls []string) {
	m := model.MakeMap[string, bool]()
	for _, url := range urls {
		m.Set(url, true)
	}
	for _, url := range s.URLs {
		m.Set(url, true)
	}
	s.URLs = m.Keys()
	s.setNodes()
}

func (s *Sitemap) setNodes() {
	n := dto.NewNode("")
	for _, url := range s.URLs {
		n.Add(url)
	}
	s.Nodes = n.Children.Values()
}

func (s *Sitemap) String() string { return util.JSON(s) }
