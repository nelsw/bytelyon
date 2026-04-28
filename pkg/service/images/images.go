package images

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/nelsw/bytelyon/pkg/https"
	"github.com/nelsw/bytelyon/pkg/s3"
	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/rs/zerolog/log"
)

func ToPublicURL(src string) (string, error) {

	if filepath.Ext(src) == ".png" {
		log.Debug().Str("src", src).Msg("already public png")
		return src, nil
	}

	if !util.IsImageFile(src) {
		return "", errors.New("invalid file format: " + src)
	}

	out, err := https.Get(src)
	if err != nil {
		return "", err
	}

	var b []byte
	if b, err = util.ToPng(out); err != nil {
		return "", err
	}

	key := strings.TrimPrefix(src, "https://")
	var url string
	if url, err = s3.PutPublicImage(key, b); err != nil {
		return "", err
	}

	log.Debug().Str("url", url).Msg("public image")

	return url, nil
}
