package service

import (
	"bytes"
	"fmt"

	"github.com/nelsw/bytelyon/pkg/client"
	"github.com/nelsw/bytelyon/pkg/db"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/repo"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
	"github.com/yuin/goldmark"
)

func SpinArticle(userID, botID, ID ulid.ULID) (err error) {

	var res *model.BotResult
	if res, err = repo.FindBotResult(userID, botID, ID, model.NewsBotType); err != nil {
		return err
	} else if link := res.GetString("data"); link != "" {
		return fmt.Errorf("article already spun: %s", link)
	} else if res.GetString("body") == "" {
		return fmt.Errorf("article body is empty")
	}

	var txt string
	txt, err = client.Prompt(
		"You are a lithium ion fire blanket salesman for a company name FireFibers",
		"Write a blog post summarizing this article: "+res.GetString("body"),
	)
	if err != nil {
		log.Err(err).Msg("failed to spin article")
		return err
	}

	log.Info().Str("text", txt).Msg("spun article")

	var buf bytes.Buffer
	if err = goldmark.Convert([]byte(txt), &buf); err != nil {
		log.Err(err).Msg("failed to convert article from md to html")
		return err
	}

	res.Data["link"], err = client.CreateArticle(
		res.ID,
		res.GetString("title"),
		buf.String(),
		res.GetString("publishedAt"),
		"https://bytelyon-public.s3.amazonaws.com/"+res.GetString("image"),
	)

	if err != nil {
		log.Error().Err(err).Msg("Failed to create article on Shopify after spinning it")
		return err
	}

	if err = db.PutItem(res); err != nil {
		log.Err(err).Msg("failed to put item and save article link")
		return err
	}

	return nil
}
