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

func Must[T any](v T, err error) T {
	Check(err)
	return v
}

func Ptr[T any](a T) *T { return &a }

func Between[T int | float64](min, max T) T {
	return T(rand.Intn(int(max)-int(min)) + int(min))
}

func Domain(s string) string {
	s = strings.TrimPrefix(s, "https://")
	s = strings.TrimPrefix(s, "http://")
	s = strings.TrimPrefix(s, "www.")
	s = strings.Split(s, "/")[0]
	for strings.Count(s, ".") > 1 {
		s = strings.Split(s, ".")[1]
	}
	return s
}

func Path(s string) string {
	if _, r, k := strings.Cut(s, Domain(s)); k {
		return r
	}
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
