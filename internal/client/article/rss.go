package article

import (
	"encoding/xml"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
)

type RSS struct {
	Channel struct {
		Items []*struct {
			URL    string `xml:"link"`
			Title  string `xml:"title"`
			Source string `xml:"source"`
			Time   *Time  `xml:"pubDate"`
		} `xml:"item"`
	} `xml:"channel"`
}

func NewRSS(url string) (*RSS, error) {
	res, err := http.Get(url)
	if err != nil {
		log.Err(err).Send()
		return nil, err
	}
	defer res.Body.Close()

	var b []byte
	if b, err = io.ReadAll(res.Body); err != nil {
		log.Err(err).Send()
		return nil, err
	}

	var rss RSS
	if err = xml.Unmarshal(b, &rss); err != nil {
		log.Err(err).Send()
		return nil, err
	}

	return &rss, nil
}
