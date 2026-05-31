package serp

import (
	"strings"

	"github.com/nelsw/bytelyon/pkg/page"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

func url(query string) string {
	return "google.com/search?q=" + strings.ReplaceAll(query, " ", "+")
}

func Find(query string, id ulid.ULID) (*Model, error) {
	return page.FindObject[*Model](url(query), id)
}

func Delete(query string, id ulid.ULID) error {
	return page.Delete(url(query), id)
}

func Create(query, content string, screenshot []byte) (m *Model, err error) {
	m = New(content)
	if err = Update(query, m); err != nil {
		log.Warn().Err(err).Msg("Failed to save serp object")
	} else if err = page.SaveContent(url(query), m.ID, content); err != nil {
		log.Warn().Err(err).Msg("Failed to save serp content")
	} else if err = page.SaveScreenshot(url(query), m.ID, screenshot); err != nil {
		log.Warn().Err(err).Msg("Failed to save serp screenshot")
	}
	return
}

func Update(query string, m *Model) (err error) {
	if err = page.SaveObject(url(query), m.ID, m); err != nil {
		log.Warn().Err(err).Msg("Failed to save serp object")
	}
	return
}
