package util

import (
	"bytes"
	"encoding/json"
	"errors"
	"image"
	"image/jpeg"
	"image/png"
	"math/rand"
	"net/http"
	"net/url"
	"path"
	"reflect"
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
	"golang.org/x/image/webp"
)

var (
	fileExtRegex = regexp.MustCompile(`.(webp|jpg|jpeg|png)`)
)

func Safe[T any](t T, err error) T {
	if err != nil {
		log.Warn().Msgf("Suppressed: %v", err)
	}
	return t
}

func Ptr[T any](t T) *T { return &t }

func Between[T int | float64](min, max T) T {
	return T(rand.Intn(int(max)-int(min)) + int(min))
}

// Domain returns the domain name from a URL in lowercase.
// Unlinke url.Parse, this ƒ does not require a protocol to determine a hostname.
func Domain(s string) string {

	s = Host(s)

	// remove subdomains
	for strings.Count(s, ".") > 1 {
		ss := strings.Split(s, ".")
		s = ss[len(ss)-2] + "." + ss[len(ss)-1]
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

func HasFileExtension(rawUrl string) bool {
	u, err := url.Parse(rawUrl)
	if err != nil {
		return false
	}
	ext := path.Ext(u.Path) // Returns ".jpg", ".pdf", etc.
	return ext != ""
}

// Host returns the host name from a URL in lowercase.
// Unlinke url.Parse, this ƒ does not require a protocol to determine a hostname.
func Host(s string) string {

	// remove protocol
	s = strings.TrimPrefix(s, "https://")
	s = strings.TrimPrefix(s, "http://")

	// remove path
	s = strings.Split(s, "/")[0]

	// remove query
	s = strings.Split(s, "?")[0]

	// remove fragment
	s = strings.Split(s, "#")[0]

	// remove port
	s = strings.Split(s, ":")[0]

	// if there is no period, it can't be a URL/URI
	if !strings.Contains(s, ".") {
		return ""
	}

	return strings.ToLower(s)
}

func JSON(a any) string {
	b, _ := json.MarshalIndent(a, "", "\t")
	return string(b)
}
