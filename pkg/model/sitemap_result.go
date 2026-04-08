package model

import (
	"maps"
	"net/url"
	"slices"
	"sort"
	"strings"

	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/oklog/ulid/v2"
)

type SitemapNode struct {
	Label string                  `json:"label"`
	Kids  []*SitemapNode          `json:"children"`
	Nodes map[string]*SitemapNode `json:"-"`
	Pages Pages                   `json:"pages"`
}

func (n *SitemapNode) SetKids() {
	if len(n.Nodes) == 0 {
		return
	}
	var children []*SitemapNode
	keys := slices.Sorted(maps.Keys(n.Nodes))
	for _, key := range keys {
		n.Nodes[key].SetKids()
		children = append(children, n.Nodes[key])
	}
	n.Kids = children
}

func newSitemapNode(label string) *SitemapNode {
	return &SitemapNode{
		Label: label,
		Kids:  []*SitemapNode{},
		Nodes: make(map[string]*SitemapNode),
		Pages: Pages{},
	}
}

type SitemapResults struct {
	BotID  ulid.ULID    `json:"botId"`
	Target string       `json:"target"`
	Domain string       `json:"domain"`
	URLs   []string     `json:"urls"`
	Pages  []Pages      `json:"pages"`
	Node   *SitemapNode `json:"node"`
}

func NewSitemapResults(results BotResults) *SitemapResults {

	// create a result pointer; define the only thing we know
	result := &SitemapResults{
		BotID:  results[0].BotID,
		Domain: util.Domain(results[0].Target),
		Target: results[0].Target,
	}

	// build a page history map to track page changes over time
	var m = make(map[string]Pages)
	for _, res := range results {
		for _, u := range res.GetStrSlice("relative") {
			p := NewPage(res.ID, u)
			p.IMG = res.GetStr("screenshot")
			p.HTML = res.GetStr("content")
			m[u] = append(m[u], p)
		}
	}

	if _, ok := m[result.Target]; !ok {
		m[result.Target] = Pages{}
	}

	// define result urls as page urls in alphabetical order
	result.URLs = slices.Collect(maps.Keys(m))
	sort.Strings(result.URLs)

	// make an array of pages for each url
	result.Pages = make([]Pages, len(result.URLs))
	for i, s := range result.URLs {
		result.Pages[i] = m[s]
	}

	// build a node tree and nodes
	node := newSitemapNode("/")

	for i, rawURL := range result.URLs {
		parsed, err := url.Parse(rawURL)
		if err != nil {
			continue
		}

		// Strip leading slash and split path into segments
		path := strings.Trim(parsed.Path, "/")
		labels := []string{parsed.Host}
		if path != "" {
			labels = append(labels, strings.Split(path, "/")...)
		}

		current := node
		for _, label := range labels {
			if _, exists := current.Nodes[label]; !exists {
				current.Nodes[label] = newSitemapNode(label)
			}
			current = current.Nodes[label]
		}
		current.Pages = result.Pages[i]
	}

	result.Node = node.Nodes[result.Domain]
	for _, key := range slices.Sorted(maps.Keys(result.Node.Nodes)) {
		result.Node.Nodes[key].SetKids()
		result.Node.Kids = append(result.Node.Kids, result.Node.Nodes[key])
	}

	return result
}
