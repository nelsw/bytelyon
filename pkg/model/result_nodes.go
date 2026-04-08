package model

import (
	"strings"
	"time"
)

type ResultNodes []*ResultNode

func (n ResultNodes) Len() int      { return len(n) }
func (n ResultNodes) Swap(i, j int) { n[i], n[j] = n[j], n[i] }
func (n ResultNodes) Less(i, j int) bool {

	// if this is a Bot Node
	if n[i].BotID == n[i].ID {
		return strings.Compare(n[i].Label, n[j].Label) == -1
	}

	// else it's a Bot Result Node
	if n[i].Type == NewsBotType {
		its, _ := time.Parse("01/02/2006", n[i].Label)
		jts, _ := time.Parse("01/02/2006", n[j].Label)
		return its.Compare(jts) == -1
	}
	return n[i].ID.Compare(n[j].ID) == -1
}

func (n ResultNodes) String() string {
	var ss = make([]string, len(n))
	for i, v := range n {
		ss[i] = v.String()
	}
	return "\n" + strings.Join(ss, ",\n")
}
