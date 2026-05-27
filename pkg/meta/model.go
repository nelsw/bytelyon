package meta

import (
	"maps"
	"slices"
	"strings"

	"github.com/nelsw/bytelyon/pkg/image"
	"github.com/nelsw/bytelyon/pkg/util"
)

type Model map[string]string

func (m Model) Title() string {
	return util.Or(
		m["title"],
		m["og:title"],
		m["twitter:title"],
	)
}

func (m Model) Image() image.Model { return image.Make(m.ImageSrc(), m.ImageAlt()) }

func (m Model) ImageSrc() string {
	return util.Or(
		m["image"],
		m["og:image"],
		m["og:image:url"],
		m["og:image:secure_url"],
		m["twitter:image"],
		m["twitter:image:src"],
	)
}

func (m Model) ImageAlt() string {
	return util.Or(
		m["og:image:alt"],
		m["twitter:image:alt"],
	)
}

func (m Model) Source() string {
	return util.Or(
		m["og:site"],
		m["og:site_name"],
		m["twitter:site"],
	)
}

func (m Model) Description() string {
	return util.Or(
		m["abstract"],
		m["description"],
		m["og:description"],
		m["twitter:description"],
	)
}

func (m Model) Keywords() []string {

	opts := []string{
		m["keywords"],
		m["news_keywords"],
		m["article:tag"],
	}

	kw := make(map[string]bool)
	for _, opt := range opts {
		if opt == "" {
			continue
		}
		kws := strings.Split(opt, ",")
		for _, w := range kws {
			kw[strings.TrimSpace(w)] = true
		}
	}

	if len(kw) == 0 {
		return []string{}
	}

	return slices.Sorted(maps.Keys(kw))
}
