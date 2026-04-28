package pages

import (
	"errors"

	"github.com/nelsw/bytelyon/internal/pw"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/repo"
	"github.com/nelsw/bytelyon/pkg/service/documents"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

func Create(url string, ctx playwright.BrowserContext) (err error) {
	var page *model.Page

	if page, err = New(url, ctx); err != nil {
		log.Err(err).Msg("new page failed")
		return
	}
	log.Info().Msg("new page succeeded")

	if err = repo.SavePage(page); err != nil {
		log.Err(err).Msg("save page failed")
	} else {
		log.Info().Msg("save page succeeded")
	}
	return
}

func New(url string, ctx playwright.BrowserContext) (page *model.Page, err error) {
	log.Debug().Str("url", url).Msg("new page")
	if page, err = NewPwPage(url, ctx); err != nil {
		log.Err(err).Str("url", url).Msg("new PW page failed")
		if page, err = NewDocumentPage(url); err != nil {
			log.Err(err).Str("url", url).Msg("new document page failed")
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
	if p, err = pw.NewPage(ctx); err != nil {
		log.Err(err).Msg("failed to create page")
		return
	}
	defer func(page playwright.Page) {
		_ = p.Close()
	}(p)

	var resp playwright.Response
	if resp, err = pw.GoTo(p, url); err != nil {
		log.Err(err).Str("url", url).Msg("failed to go to page")
		return
	} else if pw.IsRequestBlocked(resp) || pw.IsPageBlocked(p) {
		log.Warn().Str("url", url).Msg("page/request is blocked")
		return nil, errors.New("blocked")
	}

	var doc *model.Document
	if doc, err = model.ParseDocument(pw.Content(p)); err != nil {
		return nil, err
	}

	page = doc.ToPage(url, t...)
	if page.Title == "" {
		page.Title = pw.Title(p)
	}
	page.ScreenshotData = pw.Screenshot(p)

	return
}
