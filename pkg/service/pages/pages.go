package pages

import (
	"errors"

	"github.com/nelsw/bytelyon/internal/pw"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/repo"
	"github.com/oklog/ulid/v2"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

func Create(id ulid.ULID, url string, ctx playwright.BrowserContext) (err error) {

	log.Trace().Str("url", url).Msg("new document")

	var p playwright.Page
	if p, err = pw.NewPage(ctx); err != nil {
		log.Err(err).Msg("failed to create document")
		return
	}
	defer func(page playwright.Page) {
		_ = page.Close()
	}(p)

	var resp playwright.Response
	if resp, err = pw.GoTo(p, url); err != nil {
		log.Err(err).Str("url", url).Msg("failed to go to document")
		return
	} else if pw.IsRequestBlocked(resp) || pw.IsPageBlocked(p) {
		log.Warn().Str("url", url).Msg("document/request is blocked")
		return errors.New("blocked")
	}

	page := model.NewPage(id, url, pw.Title(p), pw.Content(p), pw.Screenshot(p))

	log.Debug().Msgf("new document %s", page)

	if err = repo.SavePage(page); err != nil {
		log.Err(err).Msg("save new document failed")
	} else {
		log.Info().Msg("saved new document succeeded")
	}

	return
}
