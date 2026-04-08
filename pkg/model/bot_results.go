package model

import (
	"slices"
	"sort"

	"github.com/rs/zerolog/log"
)

type BotResults []*BotResult

func (b BotResults) ToNodes() (nodes ResultNodes) {

	if len(b) == 0 {
		return
	}

	switch t := b[0].Type; t {
	case NewsBotType:
		nodes = b.ToNewsResultNodes()
	case SearchBotType:
		nodes = b.searchResultNodes()
	case SitemapBotType:
		nodes = b.sitemapResultNodes()
	default:
		log.Warn().Msgf("bot type [%s] not supported", t)
	}

	sort.Sort(nodes)
	slices.Reverse(nodes)

	return
}

func (b BotResults) ToNewsResultNodes() ResultNodes {

	var m = make(map[string]BotResults)
	for _, r := range b {
		m[r.Label()] = append(m[r.Label()], r)
	}

	var nodes ResultNodes
	for k, v := range m {
		nodes = append(nodes, &ResultNode{
			ID:     v[0].ID,
			BotID:  v[0].BotID,
			Label:  k,
			Type:   v[0].Type,
			Target: v[0].Target,
			Rows:   v,
		})
	}
	return nodes
}

func (b BotResults) searchResultNodes() ResultNodes {
	var nodes ResultNodes
	for _, r := range b {
		nodes = append(nodes, &ResultNode{
			ID:     r.ID,
			BotID:  r.BotID,
			Label:  r.Label(),
			Type:   r.Type,
			Target: r.Target,
			Rows:   r.Data["pages"],
		})
	}
	return nodes
}

func (b BotResults) sitemapResultNodes() ResultNodes {

	var nodes ResultNodes
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

		nodes = append(nodes, &ResultNode{
			ID:     r.ID,
			BotID:  r.BotID,
			Label:  r.Label(),
			Type:   r.Type,
			Target: r.Target,
			Rows:   rows,
		})
	}

	return nodes
}
