package sitemap

import (
	"encoding/json"
	"errors"

	"github.com/nelsw/bytelyon/pkg/s3"
	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/rs/zerolog/log"
)

func (m *Model) Delete() (err error) {

	//for url, ids := range m.Entries {
	//	for _, id := range ids {
	//		err = errors.Join(page.Delete(url, id))
	//	}
	//}

	if err = errors.Join(s3.Delete(m.Key(), false)); err != nil {
		log.Warn().Err(err).Msg("failed to delete sitemap")
	}

	return
}

func (m *Model) Find() (ok bool) {

	out, err := s3.Get(m.Key(), false)
	if err != nil {
		log.Warn().Err(err).Msg("failed to find sitemap")
		return
	}

	if err = json.Unmarshal(out, &m); err != nil {
		log.Warn().Err(err).Msg("failed to unmarshal sitemap")
		return
	}

	return true
}

func (m *Model) Save() {

	µ := New(m.UserID, m.Domain)
	if µ.Find() {
		µ.Merge(m)
	} else {
		µ = m
	}

	if err := s3.Put(µ.Key(), util.JSON(µ), false); err != nil {
		log.Warn().Err(err).Msg("failed to save sitemap")
	}
}
