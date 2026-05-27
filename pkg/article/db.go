package article

import (
	"github.com/nelsw/bytelyon/pkg/document"
	"github.com/nelsw/bytelyon/pkg/page"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

func Find(url string, id ulid.ULID) (Model, error) {
	return page.FindObject[Model](url, id)
}

func Create(url, title string, id ulid.ULID, content string, screenshot []byte) error {
	doc := document.New(url, content)
	a := &Model{
		Body:        doc.Paragraphs(),
		Description: doc.Meta.Description(),
		ID:          id,
		Image:       doc.Meta.Image(),
		Keywords:    doc.Meta.Keywords(),
		Source:      doc.Meta.Source(),
		Title:       title,
		URL:         url,
	}

	if err := page.SaveObject(a.URL, a.ID, a); err != nil {
		log.Warn().Err(err).Msg("failed to save article object")
		return err
	}

	if err := page.SaveScreenshot(a.URL, a.ID, screenshot); err != nil {
		log.Warn().Err(err).Msg("failed to save article screenshot")
		return err
	}

	return nil
}
