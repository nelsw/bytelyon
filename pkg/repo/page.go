package repo

import (
	"fmt"
	"strings"

	"github.com/nelsw/bytelyon/pkg/db"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/s3"
	"github.com/rs/zerolog/log"
)

func SavePage(page *model.Page) (err error) {

	l := log.With().Str("url", page.URL).Logger()
	path := fmt.Sprintf("%s/%s", strings.TrimPrefix(page.URL, "https://"), page.CreatedAt)

	if len(page.ContentData) > 0 {
		key := path + ".html"
		if err = s3.PutPrivateObject(key, []byte(page.ContentData)); err != nil {
			l.Err(err).Msg("failed to put page content")
		} else {
			l.Debug().Msgf("put page content: %s", key)
			page.ContentKey = key
		}
	}
	if len(page.ScreenshotData) > 0 {
		key := path + ".png"
		if _, err = s3.PutPublicImage(key, page.ScreenshotData); err != nil {
			l.Err(err).Msg("failed to put page screenshot")
		} else {
			l.Debug().Msgf("put page screenshot: %s", page.ScreenshotKey)
			page.ScreenshotKey = key
		}
	}

	if err = db.Put(page); err != nil {
		l.Err(err).Msg("failed to put page")
	} else {
		l.Info().Msg("put page")
	}

	return
}
