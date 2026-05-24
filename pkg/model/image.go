package model

type Images []Image

type Image map[string]string

func MakeImage(src, alt string) Image {
	return map[string]string{"src": src, "altText": alt}
}

func (i Image) GetSrc() string {
	return i["src"]
}

func (i Image) GetAlt() string {
	return i["altText"]
}

func (i Image) SetSrc(s string) {
	i["src"] = s
}

func (i Image) SetAlt(s string) {
	i["altText"] = s
}
