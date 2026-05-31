package image

import (
	"bytes"
	"errors"
	"image"
	"image/jpeg"
	"image/png"
	"net/http"
	"path/filepath"

	"github.com/nelsw/bytelyon/pkg/https"
	"github.com/nelsw/bytelyon/pkg/id"
	"github.com/nelsw/bytelyon/pkg/s3"
	"github.com/rs/zerolog/log"
	"golang.org/x/image/webp"
)

type Model struct {
	URL string `json:"url"`
	ALT string `json:"altText"`
}

func (m *Model) IsPNG() bool { return filepath.Ext(m.URL) == ".png" }

func (m *Model) ConvertToPNG() (ok bool) {

	if filepath.Ext(m.URL) == ".png" {
		return true
	}

	b, err := https.Get(m.URL)
	if err != nil {
		return
	}

	var i image.Image
	switch t := http.DetectContentType(b); t {
	case "image/png":
		i, err = png.Decode(bytes.NewReader(b))
	case "image/jpeg", "image/jpg":
		i, err = jpeg.Decode(bytes.NewReader(b))
	case "image/webp":
		i, err = webp.Decode(bytes.NewReader(b))
	default:
		err = errors.New("unsupported image type: " + t)
	}

	if err != nil {
		log.Warn().Err(err).Str("url", m.URL).Msg("failed to decode image")
		return
	}

	buf := new(bytes.Buffer)
	if err = png.Encode(buf, i); err != nil {
		log.Warn().Err(err).Str("url", m.URL).Msg("failed to encode image")
		return
	}

	key := id.NewUUID(m.URL).String() + ".png"
	if err = s3.Put(key, buf.Bytes(), true); err != nil {
		log.Warn().Err(err).Str("url", m.URL).Msg("failed to put public image")
		return
	}

	m.URL = "https://bytelyon-public.s3.amazonaws.com/" + key
	log.Debug().Str("url", m.URL).Msg("public image url")

	return true
}
