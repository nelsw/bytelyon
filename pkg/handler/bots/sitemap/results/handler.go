package bots

import (
	"maps"
	"net/http"
	"net/url"
	"slices"
	"sort"
	"strings"

	. "github.com/nelsw/bytelyon/pkg/api"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/repo"
	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/oklog/ulid/v2"
)

type Result struct {
	BotID  ulid.ULID     `json:"botId"`
	Type   model.BotType `json:"type"`
	Target string        `json:"target"`
	Domain string        `json:"domain"`
	Node   *Node         `json:"node"`
	URLs   []string      `json:"urls"`
	Pages  []model.Pages `json:"pages"`
}

func (r *Result) buildNodeTree() {

	node := &Node{
		Label: "/",
		Nodes: make(map[string]*Node),
		Pages: model.Pages{},
	}

	for i, rawURL := range r.URLs {
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
				current.Nodes[label] = &Node{
					Label: label,
					Nodes: make(map[string]*Node),
					Pages: model.Pages{},
				}
			}
			current = current.Nodes[label]
		}
		current.URL = rawURL
		current.Pages = r.Pages[i]
	}

	r.Node = node.Nodes[r.Domain]
	for _, key := range slices.Sorted(maps.Keys(r.Node.Nodes)) {
		r.Node.Nodes[key].SetChildren()
		r.Node.Children = append(r.Node.Children, r.Node.Nodes[key])
	}
}

type Node struct {
	Label    string           `json:"label"`
	Children []*Node          `json:"children"`
	Pages    model.Pages      `json:"pages"`
	URL      string           `json:"url"`
	Nodes    map[string]*Node `json:"-"`
}

func (n *Node) SetChildren() {
	if len(n.Nodes) == 0 {
		return
	}
	var children []*Node
	keys := slices.Sorted(maps.Keys(n.Nodes))
	for _, key := range keys {
		n.Nodes[key].SetChildren()
		children = append(children, n.Nodes[key])
	}
	n.Children = children
}

func Handler(r Request) (Response, error) {

	r.Log()

	switch r.Method() {
	case http.MethodGet:
		return handleGet(r), nil
	}

	return r.NI(), nil
}

func handleGet(r Request) Response {

	// find results, fail fast if empty
	results := repo.FindBotResults(r.UserID(), r.ID(), model.SitemapBotType)
	if len(results) == 0 {
		return r.NC()
	}

	// create a result pointer; define the only thing we know
	result := &Result{
		BotID:  r.ID(),
		Domain: util.Domain(results[0].Target),
		Target: results[0].Target,
		Type:   model.SitemapBotType,
	}

	// build a page history map to track page changes over time
	var m = make(map[string]model.Pages)
	for _, res := range results {
		for _, u := range res.GetStrSlice("relative") {
			p := model.NewPage(u, res.Timestamp())
			p.IMG = res.GetStr("screenshot")
			p.HTML = res.GetStr("content")
			m[u] = append(m[u], p)
		}
	}

	// define result urls as page urls in alphabetical order
	result.URLs = slices.Collect(maps.Keys(m))
	sort.Strings(result.URLs)

	// make an array of pages for each url
	result.Pages = make([]model.Pages, len(result.URLs))
	for i, s := range result.URLs {
		result.Pages[i] = m[s]
	}

	// build a node tree and nodes
	result.buildNodeTree()

	return r.OK(result)
}
