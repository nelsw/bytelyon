package article

import (
	"encoding/json"

	"github.com/nelsw/bytelyon/pkg/page"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

func (m *Model) Create() (err error) {
	if err = page.Create(m.URL, m.ID, m.screenshot, m); err != nil {
		log.Warn().Err(err).Msg("failed to create article")
	}
	return
}

func Find(url string, id ulid.ULID) (*Model, error) {

	π, err := page.Find(url, id)
	if err != nil {
		return nil, err
	}

	var µ Model
	if err = json.Unmarshal(π.Data, &µ); err != nil {
		log.Warn().Err(err).Msg("failed to unmarshal article")
		return nil, err
	}

	return &µ, nil
}
