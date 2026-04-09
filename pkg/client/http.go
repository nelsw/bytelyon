package client

import (
	"bytes"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
)

func Get(url string) ([]byte, error) {
	res, err := http.Get(url)
	if err != nil {
		log.Err(err).Str("url", url).Msg("failed to get")
		return nil, err
	}
	defer res.Body.Close()
	return io.ReadAll(res.Body)
}

func PostJSON(url string, b []byte, h map[string]string) ([]byte, error) {

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range h {
		req.Header.Set(k, v)
	}

	var res *http.Response
	if res, err = http.DefaultClient.Do(req); err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	return io.ReadAll(res.Body)
}
