package image

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"path/filepath"

	"github.com/nelsw/bytelyon/pkg/https"
	"github.com/nelsw/bytelyon/pkg/id"
	"github.com/nelsw/bytelyon/pkg/s3"
	"github.com/rs/zerolog/log"
	"golang.org/x/image/webp"
)

type Models []Model

type Model struct {
	URL string `json:"url"`
	ALT string `json:"altText"`
}

func New(url, alt string) *Model {
	return &Model{
		URL: url,
		ALT: alt,
	}
}

func (m *Model) IsPNG() bool { return filepath.Ext(m.URL) == ".png" }

func (m *Model) hasGraphicExtension() bool {
	return filepath.Ext(m.URL) == ".png" || filepath.Ext(m.URL) == ".jpg" || filepath.Ext(m.URL) == ".jpeg"
}

func (m *Model) ConvertToPNG() bool {
	ext := filepath.Ext(m.URL)
	if ext == ".png" {
		return true
	} else if ext != ".jpg" && ext != ".jpeg" && ext != ".webp" {
		return false
	}

	b, err := https.Get(m.URL)
	if err != nil {
		return false
	}

	var i image.Image
	if ext == ".webp" {
		i, err = webp.Decode(bytes.NewReader(b))
	} else {
		i, err = jpeg.Decode(bytes.NewReader(b))
	}

	if err != nil {
		log.Warn().Err(err).Str("url", m.URL).Msg("failed to decode image")
		return false
	}

	buf := new(bytes.Buffer)
	if err = png.Encode(buf, i); err != nil {
		log.Warn().Err(err).Str("url", m.URL).Msg("failed to encode image")
		return false
	}

	var publicURL string
	if publicURL, err = s3.PutPublicImage(id.NewUUID(m.URL).String(), buf.Bytes()); err != nil {
		log.Warn().Err(err).Str("url", m.URL).Msg("failed to put public image")
		return false
	}

	m.URL = publicURL
	log.Debug().Str("url", m.URL).Msg("public image url")

	return true
}
