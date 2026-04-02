package model

import (
	"sort"

	"github.com/nelsw/bytelyon/pkg/util"
)

type Bots []*Bot

func (bots Bots) ToNodes() (nodes Nodes) {
	for _, bot := range bots {
		label := bot.Target
		if bot.Type == SitemapBotType {
			label = util.Domain(bot.Target)
		}
		nodes = append(nodes, &Node{
			ID:        bot.ID,
			BotID:     bot.ID,
			Label:     label,
			Type:      bot.Type,
			Target:    bot.Target,
			Frequency: bot.Frequency,
			Lazy:      true,
		})
	}
	sort.Sort(nodes)
	return
}
