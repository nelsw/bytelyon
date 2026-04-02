package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/joho/godotenv"
	"github.com/nelsw/bytelyon/pkg/client"
	"github.com/nelsw/bytelyon/pkg/db"
	"github.com/nelsw/bytelyon/pkg/logs"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/repo"
	"github.com/nelsw/bytelyon/pkg/s3"
	"github.com/nelsw/bytelyon/pkg/util"
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
	for _, r := range findResults() {
		log.Info().Msgf("before: %+v", r)

		if r.GetStr("image") == "" || util.IsPng(r.GetStr("image")) {
			continue
		}

		b, err := client.Get(r.GetStr("image"))
		if err != nil {
			log.Err(err).Send()
			return
		}

		b, err = util.ToPng(b)
		if err != nil {
			log.Err(err).Send()
			return
		}

		bot := &model.Bot{
			UserID: r.UserID,
			Type:   model.NewsBotType,
			Target: r.Target,
		}

		var key string
		if key, err = s3.PutPublicBotData(bot, fmt.Sprintf("image/%s.png", r.ID), b); err != nil {
			log.Warn().Err(err).Msg("Failed to save news article image")
			return
		}

		r.Set("image", key)
		if err = db.PutItem(r); err != nil {
			log.Err(err).Send()
		}

		log.Info().Msgf("after: %+v", r)
	}
}

func findResult() *model.BotResult {
	r, err := repo.FindBotResult(
		users[0].ID,
		ulid.MustParse("01KMXJS68EKK50P412N58HSPSA"),
		ulid.MustParse("01KMXJSD303V5QAAXCEB1FNRJ7"),
		model.NewsBotType,
	)
	if err != nil {
		panic(err)
	}
	return r
}

func findResults() model.BotResults {
	userID := ulid.MustParse("01KM010XK0HY8HWWFPJTZGRF0F")
	botID := ulid.MustParse("01KMXJS68EKK50P412N58HSPSA")
	return repo.FindBotResults(userID, botID, model.NewsBotType)
}

func fixResult(r *model.BotResult) {

}

func fixResults(results model.BotResults) {

	for i, r := range results {
		if i == 0 {
			continue
		}
		log.Info().Msgf("before: %+v", r)
		fixImage(r)
		db.PutItem(r)
		log.Info().Msgf("after: %+v", r)
	}
}

func fixData(r *model.BotResult) {
	out, err := s3.GetPublicObject(r.GetStr("content"))
	if err != nil {
		panic(err)
	}
	var doc *model.Document
	doc, err = model.NewDocument(r.ID, string(out))
	if err != nil {
		panic(err)
	}
	log.Info().EmbedObject(doc).Msgf("%+v", doc)

	var body string
	for _, p := range doc.Paragraphs {
		if strings.Count(p, r.GetStr("source")) > 1 ||
			strings.Contains(p, "RELATED:") ||
			strings.Contains(p, "Related:") {
			continue
		}
		body += p + "\n"
	}
	r.Set("body", body)
	if v := doc.Title; v != "" {
		r.Set("title", v)
	}
	if v, k := doc.MetaImage(); k {
		r.Set("image", v)
	}
	if v, k := doc.MetaImageAlt(); k {
		r.Set("imageAlt", v)
	}
	if v, k := doc.MetaKeywords(); k {
		r.Set("keywords", v)
	}
	if v, k := doc.MetaDescription(); k {
		r.Set("description", v)
	}
	if v, k := doc.MetaSite(); k {
		r.Set("site", v)
	}
	log.Debug().Msgf("%+v", r)
}

func fixImage(r *model.BotResult) {
	src := r.GetStr("image")
	if src == "" {
		log.Warn().Msg("no news image to process")
		return
	} else if !util.IsPng(src) {
		log.Warn().Str("image", src).Msg("news image is not a png")
		return
	}

	res, err := http.Get(src)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to download news article")
		return
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	var b []byte
	if b, err = io.ReadAll(res.Body); err != nil {
		log.Warn().Err(err).Msg("Failed to read news article")
		return
	}
	bot := &model.Bot{
		UserID: r.UserID,
		Type:   model.NewsBotType,
		Target: r.Target,
	}
	var key string
	if key, err = s3.PutPublicBotData(bot, fmt.Sprintf("image/%s.png", r.ID), b); err != nil {
		log.Warn().Err(err).Msg("Failed to save news article image")
		return
	}

	r.Set("image", key)
}
