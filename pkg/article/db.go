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

func Create(url, title string, id ulid.ULID, content string, screenshot []byte) (err error) {

	doc := document.New(content)

	a := &Model{
		Body:        doc.Paragraphs(),
		Description: doc.Description(),
		ID:          id,
		Image:       doc.Image(),
		Keywords:    doc.Keywords(),
		Source:      doc.Source(),
		Title:       title,
		URL:         url,
	}

	if err = page.SaveObject(a.URL, a.ID, a); err != nil {
		log.Warn().Err(err).Msg("failed to save article object")
	} else if err = page.SaveScreenshot(a.URL, a.ID, screenshot); err != nil {
		log.Warn().Err(err).Msg("failed to save article screenshot")
	}

	return
}
