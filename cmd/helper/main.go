package main

import (
	"strings"

	"github.com/joho/godotenv"
	"github.com/nelsw/bytelyon/pkg/logs"
	"github.com/nelsw/bytelyon/pkg/manager"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/repo"
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
	userID := ulid.MustParse("01KM010XK0HY8HWWFPJTZGRF0F")
	botID := ulid.MustParse("01KMXJS68EKK50P412N58HSPSA")
	var results model.BotResults
	results = repo.FindBotResults(userID, botID, model.NewsBotType)
	//ID := ulid.MustParse("01KMXJSEDYMNE7XFRMWFQ7Q1R0")
	//result, err := repo.FindBotResult(userID, botID, ID, model.NewsBotType)
	//if err != nil {
	//	panic(err)
	//}
	//results = append(results, result)
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
