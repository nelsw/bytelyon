package dto

import (
	"encoding/json"
	"strings"

	"github.com/nelsw/bytelyon/pkg/model"
)

type Node struct {
	Children model.Map[string, *Node]
	Label    string
	Data     any
}

func NewNode(label string, data ...any) *Node {
	n := &Node{
		Children: model.MakeMap[string, *Node](),
		Label:    label,
	}
	if len(data) > 0 {
		n.Data = data[0]
	}
	return n
}

func (n *Node) MarshalJSON() ([]byte, error) {
	m := map[string]any{
		"label": n.Label,
	}
	if n.Data != nil {
		m["data"] = n.Data
	}
	if len(n.Children) > 0 {
		m["children"] = n.Children.Values()
	}
	return json.Marshal(m)
}

func (n *Node) UnmarshalJSON(b []byte) error {
	var alias struct {
		Children []*Node `json:"children"`
		Label    string  `json:"label"`
		Data     any     `json:"data"`
	}
	if err := json.Unmarshal(b, &alias); err != nil {
		return err
	}
	n.Children = model.MakeMap[string, *Node]()

	return json.Unmarshal(b, n)
}

func (n *Node) Add(url string) {
	node := n
	uri := url[8:]
	for idx := strings.Index(uri, "/"); idx != -1; idx = strings.Index(uri, "/") {
		label := uri[:idx]
		if node.Children[label] == nil {
			node.Children[label] = NewNode(label)
		}
		node = node.Children[label]
		uri = uri[idx+1:]
	}
	if _, ok := node.Children[uri]; !ok {
		node.Children[uri] = NewNode(uri)
		node.Children[uri].Data = url
	}
}
