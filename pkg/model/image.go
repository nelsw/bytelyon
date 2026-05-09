package model

type Image struct {
	URL string `json:"url"`
	ALT string `json:"altText"`
}

func MakeImage(url, alt string) Image {
	return Image{
		URL: url,
		ALT: alt,
	}
}
