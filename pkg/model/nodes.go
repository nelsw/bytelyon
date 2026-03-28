package model

import (
	"strings"
)

type Nodes []*Node

func (n Nodes) Len() int           { return len(n) }
func (n Nodes) Swap(i, j int)      { n[i], n[j] = n[j], n[i] }
func (n Nodes) Less(i, j int) bool { return n[i].ID.Compare(n[j].ID) < 0 }

func (n Nodes) String() string {
	var ss = make([]string, len(n))
	for i, v := range n {
		ss[i] = v.String()
	}
	return "\n" + strings.Join(ss, ",\n")
}
