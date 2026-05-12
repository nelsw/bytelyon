package em

import (
	"encoding/json"

	"github.com/nelsw/bytelyon/pkg/entity"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/s3"
	"github.com/nelsw/bytelyon/pkg/util"
)

func DeleteNews(bot *model.Bot, utc ...uint64) {
	if x := GetNews(bot); len(utc) > 0 {
		x.Delete(utc[0])
		s3.PutPrivateObject(bot.Key(), util.JSON(x))
	} else {
		s3.DeletePrivateObject(bot.Key())
	}
}

func GetNews(bot *model.Bot) model.Map[uint64, *entity.News] {

	out, err := s3.GetPrivateObject(bot.Key())
	if err != nil {
		return model.MakeMap[uint64, *entity.News]()
	}

	var pages model.Map[uint64, *entity.News]
	if err = json.Unmarshal(out, &pages); err != nil {
		return model.MakeMap[uint64, *entity.News]()
	}

	return pages
}
