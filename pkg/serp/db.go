package serp

import (
	"strings"

	"github.com/nelsw/bytelyon/pkg/page"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

func Find(query string, id ulid.ULID) (*Model, error) {
	return page.FindObject[*Model]("google.com/search?q="+strings.ReplaceAll(query, " ", "+"), id)
}

func Delete(query string, id ulid.ULID) error {
	return page.Delete("google.com/search?q="+strings.ReplaceAll(query, " ", "+"), id)
}

func Create(query, content string, screenshot []byte) (m *Model, err error) {
	m = New(query, content, screenshot)
	if err = page.SaveObject(m.URL, m.ID, m); err != nil {
		log.Warn().Err(err).Msg("Failed to save serp object")
	} else if err = page.SaveContent(m.URL, m.ID, m.Content); err != nil {
		log.Warn().Err(err).Msg("Failed to save serp content")
	} else if err = page.SaveScreenshot(m.URL, m.ID, m.Screenshot); err != nil {
		log.Warn().Err(err).Msg("Failed to save serp screenshot")
	}
	return
}

func Update(m *Model) (err error) {
	if err = page.SaveObject(m.URL, m.ID, m); err != nil {
		log.Warn().Err(err).Msg("Failed to save serp object")
	}
	return
}
