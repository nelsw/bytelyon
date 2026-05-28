package image

type Models []Model

type Model struct {
	URL string `json:"url"`
	ALT string `json:"altText"`
}

func Make(url, alt string) Model {
	return Model{
		URL: url,
		ALT: alt,
	}
}
