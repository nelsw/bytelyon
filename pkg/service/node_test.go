package service

import (
	"fmt"
	"testing"

	"github.com/joho/godotenv"
	"github.com/nelsw/bytelyon/pkg/logger"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"
)

func Test_News_Nodes(t *testing.T) {
	godotenv.Load()
	logger.Init()
	userID := ulid.MustParse("01KM01JC9PS1R4X4FDJNFAR4AZ")
	nodes := BotNodes(userID, model.NewsBotType)
	assert.NotEmpty(t, nodes)
	for _, node := range nodes {
		children := BotResultNodes(userID, node.BotID, node.Type)
		t.Log("node.Label", node.Label)
		for _, child := range children {
			t.Log("child.Label", child.Label)
			t.Log("child.ID", child.ID)
			t.Log("child.Type", child.Type)
			t.Log("child.Target", child.Target)
			t.Log("child.rows", child.Rows)
		}
	}

}

func Test_Search_Nodes(t *testing.T) {
	godotenv.Load()
	logger.Init()
	userID := ulid.MustParse("01KM01JC9PS1R4X4FDJNFAR4AZ")

	nodes := BotNodes(userID, model.SearchBotType)
	assert.NotEmpty(t, nodes)
	for _, node := range nodes {
		children := BotResultNodes(userID, node.BotID, node.Type)
		t.Log("node.Label", node.Label)
		for _, child := range children {
			t.Log("child.Label", child.Label)
			t.Log("child.ID", child.ID)
			t.Log("child.Type", child.Type)
			t.Log("child.Target", child.Target)
			t.Log("child.rows", child.Rows)
		}
	}
	t.Log(nodes.String())
}

func Test_Sitemap_Nodes(t *testing.T) {
	godotenv.Load()
	logger.Init()
	userID := ulid.MustParse("01KM01JC9PS1R4X4FDJNFAR4AZ")
	nodes := BotNodes(userID, model.SitemapBotType)
	assert.NotEmpty(t, nodes)
	for _, node := range nodes {
		children := BotResultNodes(userID, node.BotID, node.Type)
		t.Log("node.Label", node.Label)
		for _, child := range children {
			t.Log("child.Label", child.Label)
			t.Log("child.ID", child.ID)
			t.Log("child.Type", child.Type)
			t.Log("child.Target", child.Target)
			t.Log("child.rows", child.Rows)
			fmt.Println()
		}
	}
	t.Log(nodes.String())
}
