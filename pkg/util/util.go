package util

import (
	"bytes"
	"errors"
	"image"
	"image/jpeg"
	"image/png"
	"math/rand"
	"net/http"
	"reflect"
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
	"golang.org/x/image/webp"
)

var (
	fileExtRegex = regexp.MustCompile(`.(webp|jpg|jpeg|png)`)
)

func Check(err error) {
	if err != nil {
		panic(err)
	}
}

func Must[T any](t T, err error) T {
	Check(err)
	return t
}

func Suppress[T any](t T, err error) T {
	if err == nil {
		log.Warn().Err(err).Msg("suppressing error")
	}
	return t
}

func IsEmpty[T any](val T) bool {
	v := reflect.ValueOf(val)
	// .IsZero() handles structs, primitives, and pointers
	return !v.IsValid() || v.IsZero()
}

func Ptr[T any](t T) *T { return &t }

func Between[T int | float64](min, max T) T {
	return T(rand.Intn(int(max)-int(min)) + int(min))
}

func Domain(s string) string {
	s = strings.TrimPrefix(s, "https://")
	s = strings.TrimPrefix(s, "http://")
	s = strings.TrimPrefix(s, "www.")
	s = strings.Split(s, "/")[0]
	return s
}

func Capitalize(s string) string {
	return strings.ToUpper(s[0:1]) + s[1:]
}

func IsImageFile(s string) bool {
	return fileExtRegex.MatchString(s)
}

func ToPng(b []byte) ([]byte, error) {

	var err error
	var i image.Image

	switch t := http.DetectContentType(b); t {
	case "image/png":
		return b, nil
	case "image/jpeg", "image/jpg":
		i, err = jpeg.Decode(bytes.NewReader(b))
	case "image/webp":
		i, err = webp.Decode(bytes.NewReader(b))
	default:
		return nil, errors.New("unsupported image type: " + t)
	}

	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	if err = png.Encode(buf, i); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func PtrOrNil[T any](a T) *T {
	if v := reflect.ValueOf(a); v.IsZero() || v.IsNil() {
		return nil
	}
	return &a
}
