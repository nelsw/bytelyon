package service

import (
	"bytes"
	"context"
	"fmt"

	"github.com/PuerkitoBio/goquery"
	"github.com/nelsw/bytelyon/pkg/aws"
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
	}

	var b []byte
	b, err = client.GetObject(
		context.Background(),
		aws.S3(),
		"bytelyon-public",
		res.GetString("content"),
	)
	if err != nil {
		log.Err(err).Msg("failed to get article")
		return err
	}

	var doc *goquery.Document
	if doc, err = goquery.NewDocumentFromReader(bytes.NewReader(b)); err != nil {
		log.Err(err).Msg("failed to parse article")
		return err
	}

	var content string
	doc.Find("p").Each(func(i int, s *goquery.Selection) { content += s.Text() + "\n" })

	var txt string
	txt, err = client.Prompt(
		"You are a lithium ion fire blanket salesman for a company name FireFibers",
		"Write a blog post summarizing this article: "+content,
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

	var link string
	link, err = client.CreateArticle(
		res.ID,
		res.GetString("title"),
		buf.String(),
		res.GetString("publishedAt"),
		res.GetString("image"),
	)

	if err != nil {
		log.Error().Err(err).Msg("Failed to create article on Shopify after spinning it")
		return err
	}

	log.Info().Str("link", link).Msg("Created article")

	res.Data["link"] = link

	if err = db.PutItem(res); err != nil {
		log.Err(err).Msg("failed to put item and save article link")
		return err
	}

	return nil
}
