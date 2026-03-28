package service

import (
	"slices"
	"sort"

	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/repo"
	"github.com/oklog/ulid/v2"
)

func BotNodes(userID ulid.ULID, botType model.BotType) model.Nodes {
	var nodes model.Nodes
	for _, bot := range repo.BotsByType(userID, botType) {
		nodes = append(nodes, &model.Node{
			ID:     bot.ID,
			BotID:  bot.ID,
			Label:  bot.Label(),
			Type:   bot.Type,
			Target: bot.Target,
			Lazy:   true,
		})
	}
	sort.Sort(nodes)
	return nodes
}

func BotResultNodes(userID, botID ulid.ULID, botType model.BotType) model.Nodes {

	children := repo.
		BotResults(userID, botID, botType).
		ToNodes(botType)

	sort.Sort(children)
	slices.Reverse(children)
	return children
}
