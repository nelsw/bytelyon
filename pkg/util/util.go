package util

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"net/http"
	"path"
	"regexp"
	"strings"

	"golang.org/x/image/webp"
)

var (
	fileExtRegex = regexp.MustCompile(`.(webp|jpg|jpeg|png)`)
)

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

func JSON(a any) []byte {
	if a == nil {
		return []byte(`{}`)
	}
	switch v := a.(type) {
	case []byte:
		return v
	case string:
		return []byte(v)
	default:
		return Safe(json.Marshal(a))
	}
}

func PrettyJSON(a any) string {
	return string(Safe(json.MarshalIndent(a, "", "\t")))
}

func PrintlnPrettyJSON(a any) {
	fmt.Println(PrettyJSON(a))
}

func Path(a ...any) string {
	var arr []string
	for _, e := range a {
		arr = append(arr, fmt.Sprint(e))
	}
	return path.Join(arr...)
}

func Trunc(s string, n int) string {
	if len(s) <= n {
		return s
	}
	s = s[:n-3] + "..."
	return s
}
