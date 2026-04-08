package main

import (
	"maps"
	"slices"
	"sort"

	"github.com/joho/godotenv"
	"github.com/nelsw/bytelyon/pkg/logs"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/repo"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

var (
	users = []*model.User{
		{ID: ulid.MustParse("01KM01JC9PS1R4X4FDJNFAR4AZ"), Name: "Guest"},
		//{ID: ulid.MustParse("01KMXGBJJE2GMCA1A9EXDGF4AJ"), Name: "Stu"},
		//{ID: ulid.MustParse("01KM010XK0HY8HWWFPJTZGRF0F"), Name: "Carl"},
	}
)

func init() {
	godotenv.Load()
	logs.Init()
}

type Node struct {
	Label    string           `json:"label"`
	Children []*Node          `json:"children"`
	Pages    model.Pages      `json:"pages"`
	URL      string           `json:"-"`
	Nodes    map[string]*Node `json:"-"`
}

func (n *Node) SetChildren() {
	if len(n.Nodes) == 0 {
		return
	}
	var children []*Node
	keys := slices.Sorted(maps.Keys(n.Nodes))
	for _, key := range keys {
		n.Nodes[key].SetChildren()
		children = append(children, n.Nodes[key])
	}
	n.Children = children
}

func main() {
	doSearchBotResults()
}

func doSearchBotResults() {
	botID := ulid.MustParse("01KM13QNP8S3PEBBX9Q5MTWRHY")
	results := repo.FindBotResults(users[0].ID, botID, model.SearchBotType)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Timestamp().Compare(results[j].Timestamp()) == 1
	})
	for _, r := range results {
		log.Info().Msgf("result: %+v", r)
	}
}
