package snippet

import (
	"sort"

	"github.com/nelsw/bytelyon/pkg/page"
	"github.com/rs/zerolog/log"
)

func Find(url string) []*Model {
	out, err := page.FindObjects[*Model](url)
	if err != nil {
		log.Warn().Err(err).Msg("failed to find snippets")
		return []*Model{}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].ID.Compare(out[j].ID) > 0
	})
	return out
}
