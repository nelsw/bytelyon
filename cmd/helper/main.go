package main

import (
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
		//{ID: ulid.MustParse("01KM01JC9PS1R4X4FDJNFAR4AZ"), Name: "Guest"},
		//{ID: ulid.MustParse("01KMXGBJJE2GMCA1A9EXDGF4AJ"), Name: "Stu"},
		{ID: ulid.MustParse("01KM010XK0HY8HWWFPJTZGRF0F"), Name: "Carl"},
	}
)

func init() {
	godotenv.Load()
	logs.Init()
}

func main() {
	doSitemapBotResults()
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

func doSitemapBotResults() {
	userID := ulid.MustParse("01KM010XK0HY8HWWFPJTZGRF0F")
	botID := ulid.MustParse("01KN7Y4FFKX51990JS9YCK1TSW")
	results := repo.FindBotResults(userID, botID, model.SitemapBotType)
	log.Info().Msgf("results: %+v", results)
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
