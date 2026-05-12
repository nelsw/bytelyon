package em

import (
	"encoding/json"

	"github.com/nelsw/bytelyon/pkg/entity"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/s3"
	"github.com/nelsw/bytelyon/pkg/util"
)

func DeleteSitemap(bot *model.Bot) {
	if e, ok := GetSitemap(bot); ok {
		s3.DeletePrivateObject(e.Key())
	}
}

func GetSitemap(bot *model.Bot) (*entity.Sitemap, bool) {
	var e = new(entity.Sitemap)
	if out, err := s3.GetPrivateObject(bot.Key()); err != nil {
		return e, false
	} else if err = json.Unmarshal(out, e); err != nil {
		return e, false
	}
	return e, true
}

func PutSitemap(e *entity.Sitemap) {
	if x, ok := GetSitemap(e.Bot); ok {
		x.Merge(e)
		s3.PutPrivateObject(x.Key(), util.JSON(x))
		for _, page := range e.Values() {
			page.Save()
		}
	}
}
