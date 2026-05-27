package snippet

import (
	"github.com/nelsw/bytelyon/pkg/page"
	"github.com/rs/zerolog/log"
)

func Find(url string) []*Model {
	out, err := page.FindObjects[*Model](url)
	if err != nil {
		log.Warn().Err(err).Msg("failed to find snippets")
		return []*Model{}
	}
	return out
}
