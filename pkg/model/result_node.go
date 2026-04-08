package model

import (
	"encoding/json"
	"time"

	"github.com/oklog/ulid/v2"
)

// ResultNode is an abstract DAO used to generalize api response data
type ResultNode struct {
	ID        ulid.ULID
	BotID     ulid.ULID
	Label     string
	Lazy      bool
	Children  ResultNodes
	Type      BotType
	Target    string
	Rows      any
	Frequency time.Duration
	Blacklist []string
}

func (n *ResultNode) String() string {
	b, _ := json.MarshalIndent(n, "", "\t")
	return string(b)
}

func (n *ResultNode) MarshalJSON() ([]byte, error) {
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
