package file

import "path"

type Ext string

const (
	HTML Ext = ".html"
	PNG      = ".png"
	JPG      = ".jpg"
	ICO      = ".ico"
	JSON     = ".json"
	JPEG     = ".jpeg"
	PDF      = ".pdf"
	WEBP     = ".webp"
)

func Extension(s string) (string, bool) {
	return path.Ext(s), true
}
