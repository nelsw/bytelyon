package search

import (
	"encoding/json"
	"errors"

	"github.com/nelsw/bytelyon/pkg/s3"
	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/rs/zerolog/log"
)

func (m *Model) Delete() (err error) {

	//for id, urls := range m.Entries {
	//	for _, url := range urls {
	//		err = errors.Join(page.Delete(url, id))
	//	}
	//}

	if err = errors.Join(s3.Delete(m.Key(), false)); err != nil {
		log.Warn().Err(err).Msg("failed to delete search")
	}

	return
}

func (m *Model) Find() (model *Model) {

	out, err := s3.Get(m.Key(), false)
	if err != nil {
		log.Warn().Err(err).Msg("failed to find search")
		return nil
	}

	if err = json.Unmarshal(out, &model); err != nil {
		log.Warn().Err(err).Msg("failed to unmarshal search")
		return nil
	}

	return
}

func (m *Model) Save() {

	µ := New(m.UserID, m.Query).Find()
	if µ != nil {
		µ.Merge(m)
	} else {
		µ = m
	}

	if err := s3.Put(µ.Key(), util.JSON(µ), false); err != nil {
		log.Warn().Err(err).Msg("failed to save search")
	}
}
