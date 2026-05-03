package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/nelsw/bytelyon/pkg/util"
)

type Sitemap struct {
	Domain    string  `json:"domain"`
	URLs      *Set    `json:"urls"`
	CreatedAt *Time   `json:"created_at"`
	UpdatedAt *Time   `json:"updated_at"`
	Nodes     []*Node `json:"nodes"`
}

func (s *Sitemap) String() string {
	b, _ := json.MarshalIndent(s, "", "\t")
	return string(b)
}

func (s *Sitemap) Get() *dynamodb.GetItemInput {
	return &dynamodb.GetItemInput{
		TableName: util.Ptr("ByteLyon_Sitemap"),
		Key: map[string]types.AttributeValue{
			"domain": &types.AttributeValueMemberS{Value: s.Domain},
		},
	}
}

func (s *Sitemap) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	s.UpdatedAt = Now()
	return &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
		"domain":    &types.AttributeValueMemberS{Value: s.Domain},
		"urls":      s.URLs.ToAttributeValue(),
		"createdAt": s.CreatedAt.ToAttributeValue(),
		"updatedAt": s.UpdatedAt.ToAttributeValue(),
	}}, nil
}

func (s *Sitemap) UnmarshalDynamoDBAttributeValue(v types.AttributeValue) (err error) {
	var m map[string]types.AttributeValue
	if m = v.(*types.AttributeValueMemberM).Value; m == nil {
		return errors.New("sitemap unmarshal value was nil")
	} else if s.CreatedAt, err = ParseTime(m["createdAt"]); err != nil {
		return fmt.Errorf("failed to parse createdAt: %w", err)
	} else if s.UpdatedAt, err = ParseTime(m["updatedAt"]); err != nil {
		return fmt.Errorf("failed to parse updatedAt: %w", err)
	}
	s.Domain = m["domain"].(*types.AttributeValueMemberS).Value
	s.URLs = ParseSet(m["urls"])
	return
}

func (s *Sitemap) SetNodes() {

	set := NewSet()
	for _, url := range s.URLs.Slice() {
		var label string
		for i, part := range strings.Split(url[8:], "/") {
			if i > 0 {
				label += "/"
			}
			label += part
			set.Add(label)
		}
	}

	var nodes = make(map[int][]*Node)
	for _, k := range set.Slice() {
		depth := strings.Count(k, "/")
		if _, ok := nodes[depth]; !ok {
			nodes[depth] = make([]*Node, 0)
		}
		nodes[depth] = append(nodes[depth], &Node{
			Depth:    depth,
			Label:    strings.Split(k, "/")[depth],
			Children: make([]*Node, 0),
			URL:      "https://" + k,
		})
	}

	var ƒ func(root *Node, nodes map[int][]*Node)

	ƒ = func(root *Node, nodes map[int][]*Node) {
		for _, node := range nodes[root.Depth+1] {
			if node.URL == root.URL+"/"+node.Label {
				root.Children = append(root.Children, node)
				ƒ(node, nodes)
			}
			if !s.URLs.Has(node.URL) {
				node.URL = ""
			}
		}
		if !s.URLs.Has(root.URL) {
			root.URL = ""
		}
	}

	ƒ(nodes[0][0], nodes)

	s.Nodes = nodes[0]
}

func NewSitemap(domain string) *Sitemap {
	return &Sitemap{
		Domain:    domain,
		URLs:      NewSet(),
		CreatedAt: Now(),
	}
}
