package model

type PageDTO struct {
	URL   string `json:"url"`
	Title string `json:"title"`
	IMG   string `json:"img"`
	HTML  string `json:"html"`
	SERP  Serp   `json:"serp,omitempty"`
}
