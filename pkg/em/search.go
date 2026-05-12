package em

import (
	"encoding/json"

	"github.com/nelsw/bytelyon/pkg/entity"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/s3"
	"github.com/nelsw/bytelyon/pkg/util"
)

func GetSearches(bot *model.Bot) model.Map[uint64, *entity.Search] {

	out, err := s3.GetPrivateObject(bot.Key())
	if err != nil {
		return model.MakeMap[uint64, *entity.Search]()
	}

	var pages model.Map[uint64, *entity.Search]
	if err = json.Unmarshal(out, &pages); err != nil {
		return model.MakeMap[uint64, *entity.Search]()
	}

	return pages
}

func PutSearch(e *entity.Search) {
	x := GetSearches(e.Bot)
	x.Set(e.UTC, e)
	s3.PutPrivateObject(e.Key(), util.JSON(x))
}
