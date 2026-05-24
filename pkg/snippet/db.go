package snippet

import (
	"github.com/nelsw/bytelyon/pkg/page"
	"github.com/rs/zerolog/log"
)

func (m *Model) Create() (err error) {
	if err = page.Create(m.URL, m.ID, m.content, m.screenshot, m); err != nil {
		log.Warn().Err(err).Msg("failed to create snippet")
	}
	return
}
