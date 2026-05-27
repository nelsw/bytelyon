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
	"net/url"
	"path"
	"regexp"
	"strings"

	"golang.org/x/image/webp"
)

var (
	fileExtRegex = regexp.MustCompile(`.(webp|jpg|jpeg|png)`)
)

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

	s = strings.TrimPrefix(s, "https://")

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

func Chomp(in, a, z string) (string, string) {

	aIdx, zIdx := strings.Index(in, a), strings.Index(in, z)
	if aIdx == -1 || zIdx == -1 {
		return "", in
	}

	out := in[aIdx+len(a) : zIdx]

	in = in[:aIdx] + in[zIdx+len(z):]

	return out, in
}

func Chomps(s, a, z string) (arr []string) {
	var res string
	for res, s = Chomp(s, a, z); res != ""; res, s = Chomp(s, a, z) {
		arr = append(arr, res)
	}
	return
}
