package fetch

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
)

type Client interface {
	Response() (*http.Response, error)
	Reader() (io.ReadCloser, error)
	Bytes() ([]byte, error)
	JSON(a any) error
	XML(a any) error
}

type client struct {
	url string
}

func New(url string) Client {
	return &client{url: url}
}

func (c client) Response() (*http.Response, error) {
	return http.Get(c.url)
}

func (c client) Reader() (io.ReadCloser, error) {
	res, err := c.Response()
	if err != nil {
		log.Err(err).Send()
		return nil, err
	}
	return res.Body, nil
}

func (c client) Bytes() ([]byte, error) {
	body, err := c.Reader()
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var b []byte
	if b, err = io.ReadAll(body); err != nil {
		log.Err(err).Send()
		return nil, err
	}
	return b, nil
}

func (c client) JSON(a any) error {
	b, err := c.Bytes()
	if err != nil {
		return err
	}
	if err = json.Unmarshal(b, &a); err != nil {
		log.Err(err).Send()
		return err
	}
	return nil
}

func (c client) XML(a any) error {
	b, err := c.Bytes()
	if err != nil {
		return err
	}
	if err = xml.Unmarshal(b, &a); err != nil {
		log.Err(err).Send()
		return err
	}
	return nil
}
