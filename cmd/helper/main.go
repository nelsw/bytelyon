package main

import (
	"encoding/json"
	"fmt"
	"maps"
	"net/url"
	"slices"
	"sort"
	"strings"

	"github.com/joho/godotenv"
	"github.com/nelsw/bytelyon/pkg/client"
	"github.com/nelsw/bytelyon/pkg/db"
	"github.com/nelsw/bytelyon/pkg/logs"
	"github.com/nelsw/bytelyon/pkg/manager"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/pw"
	"github.com/nelsw/bytelyon/pkg/repo"
	. "github.com/nelsw/bytelyon/pkg/util"
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

type Result struct {
	BotID  string        `json:"botId"`
	Target string        `json:"target"`
	Type   model.BotType `json:"type"`
	Domain string        `json:"domain"`
	URLs   []string      `json:"urls"`
	Pages  []model.Pages `json:"pages"`
	Node   *Node         `json:"-"`
	Nodes  []*Node       `json:"nodes"`
}

func (r *Result) BuildNodeTree() {

	node := &Node{
		Label: "/",
		Nodes: make(map[string]*Node),
		Pages: model.Pages{},
	}

	for i, rawURL := range r.URLs {
		parsed, err := url.Parse(rawURL)
		if err != nil {
			continue
		}

		// Strip leading slash and split path into segments
		path := strings.Trim(parsed.Path, "/")
		labels := []string{parsed.Host}
		if path != "" {
			labels = append(labels, strings.Split(path, "/")...)
		}

		current := node
		for _, label := range labels {
			if _, exists := current.Nodes[label]; !exists {
				current.Nodes[label] = &Node{
					Label: label,
					Nodes: make(map[string]*Node),
					Pages: model.Pages{},
				}
			}
			current = current.Nodes[label]
		}
		current.URL = rawURL
		current.Pages = r.Pages[i]
	}

	r.Node = node.Nodes[r.Domain]
}

func (r *Result) BuildNodes() {
	if r.Node == nil || len(r.Node.Nodes) == 0 {
		return
	}

	node := &Node{
		Label: r.Domain,
		URL:   r.Node.URL,
		Pages: r.Node.Pages,
	}

	keys := slices.Sorted(maps.Keys(r.Node.Nodes))
	for _, key := range keys {
		r.Node.Nodes[key].SetChildren()
		node.Children = append(node.Children, r.Node.Nodes[key])
	}

	r.Nodes = []*Node{node}
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

	results := doSitemapBotResults()
	if len(results) == 0 {
		return
	}

	var r = Result{
		Domain: Domain(results[0].Target),
	}

	var m = make(map[string]model.Pages)
	for _, result := range results {
		for _, u := range result.GetStrSlice("relative") {
			p := model.NewPage(u, result.Timestamp())
			p.IMG = result.GetStr("screenshot")
			p.HTML = result.GetStr("content")
			m[u] = append(m[u], p)
		}
	}
	r.URLs = slices.Collect(maps.Keys(m))
	r.Pages = make([]model.Pages, len(r.URLs))

	sort.Strings(r.URLs)

	for i, s := range r.URLs {
		r.Pages[i] = m[s]
	}

	r.BuildNodeTree()
	r.BuildNodes()
	b, _ := json.MarshalIndent(r.Nodes, "", "  ")
	fmt.Println(string(b))
}

func doStuff() {
	userID := ulid.MustParse("01KM010XK0HY8HWWFPJTZGRF0F")
	botID := ulid.MustParse("01KN7Q27G4MA3D75A1FJEE77QE")
	var results model.BotResults
	//results = repo.FindBotResults(userID, botID, model.SearchBotType)
	ID := ulid.MustParse("01KN7Q27G4MA3D75A1FJEE77QE")
	result, err := repo.FindBotResult(userID, botID, ID, model.SearchBotType)
	if err != nil {
		panic(err)
	}
	results = append(results, result)
	for _, r := range results {
		log.Info().Msgf("before: %+v", r)
		job := manager.NewJob(&model.Bot{
			UserID: r.UserID,
			Type:   model.NewsBotType,
			Target: r.Target,
		})
		var body []string
		for _, p := range strings.Split(r.GetStr("body"), "\n") {
			body = append(body, p)
		}
		r.Set("body", body)
		job.UpdateNewsResult(r)

		log.Info().Msgf("after: %+v", r)
	}
}

func doSitemapBotResults() model.BotResults {
	botID := ulid.MustParse("01KM19AGWKT3TWD1JX09KWB5KF")
	results := repo.FindBotResults(users[0].ID, botID, model.SitemapBotType)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Timestamp().Compare(results[j].Timestamp()) == 1
	})
	return results
}

func doSearchBotResult() {
	pw.Init()
	userID := ulid.MustParse("01KM010XK0HY8HWWFPJTZGRF0F")
	//botID := ulid.MustParse("01KN7Q27G4MA3D75A1FJEE77QE")
	//ID := ulid.MustParse("01KN7Q27G4MA3D75A1FJEE77QE")

	bot := Must(repo.FindBot(userID, "ev fire blankets for sale", model.SearchBotType))
	log.Info().Msgf("bot: %+v", bot)

	bro := Must(pw.NewBrowser(bot.Headless))
	defer bro.Close()
	ctx := Must(client.NewContext(bro, bot.Fingerprint.GetState()))
	defer ctx.Close()

	job := manager.NewJob(bot, ctx)

	job.Work()

	if state, err := ctx.StorageState(); err != nil {
		log.Warn().Err(err).Msg("Failed to get storage state")
	} else {
		bot.Fingerprint.SetState(state)
	}

	Check(db.PutItem(bot))
}
