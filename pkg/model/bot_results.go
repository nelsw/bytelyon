package model

import (
	"maps"
	"slices"
	"time"

	"github.com/rs/zerolog/log"
)

type BotResults []*BotResult

func (b BotResults) ToNodes(t BotType) Nodes {
	switch t {
	case NewsBotType:
		return b.ToNewsResultNodes()
	case SearchBotType:
		return b.searchResultNodes()
	case SitemapBotType:
		return b.sitemapResultNodes()
	}
	log.Warn().Msgf("bot type [%s] not supported", t)
	return Nodes{}
}

func (b BotResults) ToNewsResultNodes() Nodes {

	var m = make(map[int64]BotResults)
	for _, r := range b {
		ts := r.Timestamp().Truncate(24 * time.Hour).Unix()
		m[ts] = append(m[ts], r)
	}

	var nodes Nodes
	for _, k := range slices.Sorted(maps.Keys(m)) {
		r := m[k][0]
		var children Nodes
		for _, v := range m[k] {
			children = append(children, &Node{
				BotID:  v.BotID,
				Label:  v.Label(),
				Type:   v.Type,
				Target: v.Target,
				Rows:   v.Data,
			})
		}
		nodes = append(nodes, &Node{
			ID:     r.ID,
			BotID:  r.BotID,
			Label:  r.Label(),
			Rows:   m[k],
			Type:   r.Type,
			Target: r.Target,
		})
	}

	return nodes
}

func (b BotResults) searchResultNodes() Nodes {
	var nodes Nodes
	for _, r := range b {
		nodes = append(nodes, &Node{
			ID:     r.ID,
			BotID:  r.BotID,
			Label:  r.Label(),
			Type:   r.Type,
			Target: r.Target,
			Rows:   r.Data["pages"].([]any),
		})
	}
	return nodes
}

func (b BotResults) sitemapResultNodes() Nodes {

	var nodes Nodes
	for _, r := range b {

		var rows []any

		s, _ := r.Data["relative"].([]any)
		for _, u := range s {
			if u != r.Target {
				rows = append(rows, map[string]any{
					"url":        u,
					"isExternal": false,
				})
			}
		}

		s, _ = r.Data["remote"].([]any)
		for _, u := range s {
			rows = append(rows, map[string]any{
				"url":        u,
				"isExternal": true,
			})
		}

		nodes = append(nodes, &Node{
			ID:       r.ID,
			BotID:    r.BotID,
			Label:    r.Label(),
			Type:     r.Type,
			Target:   r.Target,
			Rows:     rows,
			Children: Nodes{},
		})
	}

	return nodes
}
