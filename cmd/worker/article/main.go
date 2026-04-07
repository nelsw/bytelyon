package main

import (
	"errors"
	"flag"
	"fmt"
	"strings"

	"github.com/joho/godotenv"
	"github.com/nelsw/bytelyon/pkg/client"
	"github.com/nelsw/bytelyon/pkg/logs"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/pw"
	"github.com/nelsw/bytelyon/pkg/service"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

func init() {
	fmt.Println(logs.CyanIntense + `
██████╗ ██╗   ██╗████████╗███████╗██╗  ██╗   ██╗ ██████╗ ███╗   ██╗
██╔══██╗╚██╗ ██╔╝╚══██╔══╝██╔════╝██║  ╚██╗ ██╔╝██╔═══██╗████╗  ██║
██████╔╝ ╚████╔╝    ██║   █████╗  ██║   ╚████╔╝ ██║   ██║██╔██╗ ██║
██╔══██╗  ╚██╔╝     ██║   ██╔══╝  ██║    ╚██╔╝  ██║   ██║██║╚██╗██║
██████╔╝   ██║      ██║   ███████╗███████╗██║   ╚██████╔╝██║ ╚████║
╚═════╝    ╚═╝      ╚═╝   ╚══════╝╚══════╝╚═╝    ╚═════╝ ╚═╝  ╚═══╝
-------------------------------------------------------------------
Article Worker` + logs.Default)
	godotenv.Load()
	logs.Init()
	pw.Init()
}

func main() {

	var u string
	var prompt string
	flag.StringVar(&u, "u", "", "article url")
	flag.StringVar(&prompt, "p", "You are a lithium ion fire blanket salesman for a company named FireFibers", "prompt")
	flag.Parse()

	title, content, err := fetchContent(u)
	if err != nil {
		log.Err(err).Msg("failed to fetch content")
		return
	}

	var a *model.Article
	if a, err = buildArticle(title, content, prompt); err != nil {
		log.Err(err).Msg("failed to build article")
		return
	}

	var link string
	if link, err = service.SpinArticle(a); err != nil {
		log.Err(err).Msg("failed to spin article")
		return
	}
	log.Info().Str("link", link).Msg("article spun")
}

func fetchContent(s string) (title, content string, err error) {

	defer pw.Client.Stop()

	var bro playwright.Browser
	if bro, err = pw.NewBrowser(false); err != nil {
		log.Err(err).Msg("failed to create browser")
		return
	}
	defer bro.Close()

	var ctx playwright.BrowserContext
	if ctx, err = client.NewContext(bro, nil); err != nil {
		log.Err(err).Msg("failed to create context")
		return
	}
	defer ctx.Close()

	var page playwright.Page

	if page, err = client.NewPage(ctx); err != nil {
		log.Err(err).Msg("failed to create new page")
		return
	}
	defer func(page playwright.Page) {
		_ = page.Close()
	}(page)

	var resp playwright.Response
	if resp, err = client.GoTo(page, s); err != nil {
		log.Err(err).Str("url", s).Msg("failed to go to page url")
	} else if client.IsRequestBlocked(resp) {
		log.Warn().
			Str("url", s).
			Int("code", resp.Status()).
			Str("reason", resp.StatusText()).
			Msg("request is blocked")
		err = errors.New("request is blocked")
		return
	} else if content, err = page.Content(); err != nil {
		log.Warn().Err(err).Msg("failed to get page content")
	} else if title, err = page.Title(); err != nil {
		log.Warn().Err(err).Msg("failed to get page title")
	} else {
		log.Info().Str("url", s).Msg("got page content")
	}
	return
}

func buildArticle(title, content, prompt string) (*model.Article, error) {

	doc, err := model.NewDocument(model.NewULID(), content)
	if err != nil {
		log.Err(err).Msg("failed to create document")
		return nil, err
	}
	log.Info().Msgf("parsed document: %s", doc)

	a := &model.Article{
		ID:     model.NewULID(),
		Title:  title,
		Prompt: prompt,
		Image:  map[string]string{},
	}

	if v, k := doc.MetaTitle(); k && a.Title == "" {
		a.Title = v
	}
	if v, k := doc.MetaDescription(); k && a.Summary == "" {
		a.Summary = v
	}
	if v, k := doc.MetaImage(); k {
		a.Image["url"] = v
	}
	if v, k := doc.MetaImageAlt(); k {
		a.Image["altText"] = v
	}

	var body []string
	for _, p := range doc.Paragraphs {
		if strings.Contains(strings.ToLower(p), "related:") {
			continue
		}
		body = append(body, p)
	}
	a.Body = strings.Join(body, "\n\n")

	return a, nil
}
