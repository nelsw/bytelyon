package model

import (
	"sort"
)

type Bots []*Bot

func (bots Bots) ToNodes() (nodes Nodes) {
	for _, bot := range bots {
		nodes = append(nodes, &Node{
			ID:     bot.ID,
			BotID:  bot.ID,
			Label:  bot.Target,
			Type:   bot.Type,
			Target: bot.Target,
			Lazy:   true,
		})
	}
	sort.Sort(nodes)
	return
}
