package snippet

import (
	"strings"

	"github.com/nelsw/bytelyon/pkg/document"
	"github.com/nelsw/bytelyon/pkg/meta"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/urls"
	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/oklog/ulid/v2"
)

type Model struct {
	Domain     string     `json:"-"`
	ID         ulid.ULID  `json:"id"`
	Links      []string   `json:"-"`
	Meta       meta.Model `json:"meta"`
	Screenshot []byte     `json:"-"`
	Title      string     `json:"title"`
	URL        string     `json:"url"`
}

func New(id ulid.ULID, url, content string, screenshot []byte) *Model {
	doc := document.New(content)
	return &Model{
		Domain:     urls.Domain(url),
		ID:         id,
		Links:      doc.Links,
		Meta:       doc.Meta,
		Screenshot: screenshot,
		Title:      util.Or(doc.Title, doc.Meta.Title()),
		URL:        urls.PR(url),
	}
}

func (m *Model) URLs() []string {
	set := model.NewSet[string]()
	for _, link := range model.NewSet(m.Links...).Slice() {
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
			set.Add("https://" + m.Domain + link)
			continue
		}

		// if the link is a url; check the host equals our domain
		if host := urls.Host(link); host != "" && host != m.Domain {
			continue
		}

		// if the link is a secure URL
		if strings.HasPrefix(link, "https://"+m.Domain) {
			set.Add(link)
			continue
		}

		// if the link is missing URL protocol
		if strings.HasPrefix(link, m.Domain) {
			set.Add("https://" + link)
			continue
		}

		// else the link is relative to this url
		if l, _, ok := strings.Cut(link, "/"); ok {
			set.Add(m.URL + "/" + l + "/" + link)
		} else {
			set.Add(m.URL + "/" + link)
		}
	}
	return set.Slice()
}
