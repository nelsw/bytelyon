package sitemaps

import (
	"maps"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/nelsw/bytelyon/pkg/api"
	"github.com/nelsw/bytelyon/pkg/db"
	"github.com/nelsw/bytelyon/pkg/model"
)

type Node struct {
	URL      string           `json:"url"`
	Label    string           `json:"label"`
	Children []*Node          `json:"children,omitempty"`
	Nodes    map[string]*Node `json:"-"`
}

func (n *Node) SetChildren() {
	if len(n.Nodes) == 0 {
		return
	}
	var children []*Node
	for _, key := range slices.Sorted(maps.Keys(n.Nodes)) {
		n.Nodes[key].SetChildren()
		children = append(children, n.Nodes[key])
	}
	if len(children) > 0 {
		n.Children = children
	}
}

func Handler(r api.Request) api.Response {
	switch r.Method() {
	case http.MethodGet:
		return handleGet(r)
	}
	return r.NI()
}

func handleGet(r api.Request) api.Response {
	m, err := db.Get(&model.Sitemap{Domain: r.Domain()})
	if err != nil {
		return r.BAD(err)
	}

	node := &Node{
		URL:   "",
		Label: "/",
		Nodes: make(map[string]*Node),
	}

	var u *url.URL
	for _, s := range m.URLs.Slice() {
		u, err = url.Parse(s)
		if err != nil {
			continue
		}

		// Strip leading slash and split path into segments
		path := strings.Trim(u.Path, "/")
		labels := []string{u.Host}
		if path != "" {
			labels = append(labels, strings.Split(path, "/")...)
		}

		current := node
		for _, label := range labels {
			if _, exists := current.Nodes[label]; !exists {
				current.Nodes[label] = &Node{
					URL:   s,
					Label: label,
					Nodes: make(map[string]*Node),
				}
			}
			current = current.Nodes[label]
		}
	}

	node = node.Nodes[r.Domain()]
	for _, key := range slices.Sorted(maps.Keys(node.Nodes)) {
		node.Nodes[key].SetChildren()
		node.Children = append(node.Children, node.Nodes[key])
	}

	return r.OK(node)
}
