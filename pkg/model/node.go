package model

import (
	"encoding/json"
	"time"

	"github.com/oklog/ulid/v2"
)

type Node struct {
	ID        ulid.ULID
	BotID     ulid.ULID
	Label     string
	Lazy      bool
	Children  Nodes
	Type      BotType
	Target    string
	Rows      any
	Frequency time.Duration
	Blacklist []string
}

func (n *Node) String() string {
	b, _ := json.MarshalIndent(n, "", "\t")
	return string(b)
}

func (n *Node) MarshalJSON() ([]byte, error) {
	m := map[string]any{
		"id":        n.ID,
		"botId":     n.BotID,
		"label":     n.Label,
		"lazy":      n.Lazy,
		"type":      n.Type,
		"target":    n.Target,
		"rows":      n.Rows,
		"frequency": n.Frequency.Nanoseconds(),
		"blacklist": n.Blacklist,
	}
	if len(n.Children) > 0 {
		m["children"] = n.Children
	}
	return json.Marshal(m)
}
