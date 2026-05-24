package news

import (
	"encoding/json"

	"github.com/nelsw/bytelyon/pkg/s3"
	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/rs/zerolog/log"
)

func (m *Model) Find() (model *Model) {

	out, err := s3.Get(m.Key(), false)
	if err != nil {
		log.Warn().Err(err).Msg("failed to get news")
		return nil
	}

	if err = json.Unmarshal(out, &model); err != nil {
		log.Warn().Err(err).Msg("failed to unmarshal news")
		return nil
	}

	return
}

func (m *Model) Save() {
	if err := s3.Put(m.Key(), util.JSON(m), false); err != nil {
		log.Warn().Err(err).Msg("failed to save news")
	}
}
