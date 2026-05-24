package model

import (
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/nelsw/bytelyon/pkg/util"
)

type Meta map[string]string

func (m Meta) Title() string {
	return util.Or(
		m["title"],
		m["og:title"],
		m["twitter:title"],
	)
}

func (m Meta) Image() Image {
	return MakeImage(m.ImageSrc(), m.ImageAlt())
}

func (m Meta) ImageSrc() string {
	return util.Or(
		m["image"],
		m["og:image"],
		m["og:image:url"],
		m["og:image:secure_url"],
		m["twitter:image"],
		m["twitter:image:src"],
	)
}

func (m Meta) ImageAlt() string {
	return util.Or(
		m["og:image:alt"],
		m["twitter:image:alt"],
	)
}

func (m Meta) Source() string {
	return util.Or(
		m["og:site"],
		m["og:site_name"],
		m["twitter:site"],
	)
}

func (m Meta) Description() string {
	return util.Or(
		m["abstract"],
		m["description"],
		m["og:description"],
		m["twitter:description"],
	)
}

func (m Meta) Keywords() []string {

	opts := []string{
		m["keywords"],
		m["news_keywords"],
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

func (m Meta) PublishedTime() string {
	return util.Or(
		m["article:published_time"],
		m["og:article:published_time"],
		m["article:modified_time"],
		m["og:article:modified_time"],
	)
}

func (m Meta) PublishedAt() time.Time {
	var t time.Time
	t, _ = time.Parse(time.RFC3339, m.PublishedTime())
	return t
}
