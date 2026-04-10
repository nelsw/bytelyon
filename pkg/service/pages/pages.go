package pages

import (
	"errors"

	"github.com/nelsw/bytelyon/pkg/client"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/repo"
	"github.com/nelsw/bytelyon/pkg/service/documents"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

func Create(url string, ctx playwright.BrowserContext) error {
	page, err := New(url, ctx)
	if err != nil {
		return err
	}
	return repo.SavePage(page)
}

func New(url string, ctx playwright.BrowserContext) (page *model.Page, err error) {
	log.Debug().Str("url", url).Msg("new page")
	if page, err = NewPwPage(url, ctx); err != nil {
		log.Err(err).Str("url", url).Msg("failed to create dynamic page")
		if page, err = NewDocumentPage(url); err != nil {
			log.Err(err).Str("url", url).Msg("failed to create static page")
		}
	}
	return
}

func NewDocumentPage(url string, t ...*model.Time) (page *model.Page, err error) {
	var doc *model.Document
	if doc, err = documents.New(url); err != nil {
		log.Err(err).Str("url", url).Msg("failed to create page")
		return
	}
	return doc.ToPage(url, t...), nil
}

func NewPwPage(url string, ctx playwright.BrowserContext, t ...*model.Time) (page *model.Page, err error) {

	var p playwright.Page
	if p, err = client.NewPage(ctx); err != nil {
		log.Err(err).Msg("failed to create page")
		return
	}
	defer func(page playwright.Page) {
		_ = p.Close()
	}(p)

	var resp playwright.Response
	if resp, err = client.GoTo(p, url); err != nil {
		log.Err(err).Str("url", url).Msg("failed to go to page")
		return
	} else if client.IsRequestBlocked(resp) || client.IsPageBlocked(p) {
		log.Warn().Str("url", url).Msg("page/request is blocked")
		return nil, errors.New("blocked")
	}

	var doc *model.Document
	if doc, err = model.ParseDocument(client.Content(p)); err != nil {
		return nil, err
	}

	page = doc.ToPage(url, t...)
	if page.Title == "" {
		page.Title = client.Title(p)
	}
	page.ScreenshotData = client.Screenshot(p)

	return
}
