package news

import (
	"encoding/json"
	"errors"
	"sort"

	"github.com/nelsw/bytelyon/pkg/article"
	"github.com/nelsw/bytelyon/pkg/s3"
	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/rs/zerolog/log"
)

func (m *Model) Delete() (err error) {

	//for url, id := range m.Entries {
	//	err = errors.Join(page.Delete(url, id))
	//}

	if err = errors.Join(s3.Delete(m.Key(), false)); err != nil {
		log.Warn().Err(err).Msg("failed to delete news")
	}

	return
}

func (m *Model) Find(eager ...bool) (ok bool) {

	out, err := s3.Get(m.Key(), false)
	if err != nil {
		log.Warn().Err(err).Msg("failed to find news")
		return
	}

	if err = json.Unmarshal(out, &m); err != nil {
		log.Warn().Err(err).Msg("failed to unmarshal news")
		return
	}

	if len(eager) > 0 {
		for url, id := range m.Entries {
			if å, _ := article.Find(url, id); å != nil {
				m.Articles = append(m.Articles, å)
			}
		}
		sort.Sort(sort.Reverse(m.Articles))
	}

	return true
}

func (m *Model) Save() {

	µ := New(m.UserID, m.Topic)
	if µ.Find() {
		µ.Merge(m)
	} else {
		µ = m
	}

	if err := s3.Put(µ.Key(), util.JSON(µ), false); err != nil {
		log.Warn().Err(err).Msg("failed to save news")
	}
}
