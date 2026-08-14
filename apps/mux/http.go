package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog/log"
)

var _client = http.Client{Timeout: 10 * time.Second}

func Get[T any](url string) (t T) {
	if out, err := Do(http.MethodGet, url, nil); err != nil {
		log.Err(err).Msg("failed to execute request")
	} else if err = json.Unmarshal(out, &t); err != nil {
		log.Err(err).Msg("failed to unmarshal response")
	}
	return
}

func Post(url string, body any) {
	if _, err := Do(http.MethodPost, url, body); err != nil {
		log.Err(err).Msg("failed to execute request")
	}
}

func Do(method, url string, body any) (b []byte, err error) {

	l := log.With().Str("method", method).Str("url", url).Logger()

	var buf io.Reader
	if body != nil {
		if b, err = json.Marshal(body); err != nil {
			l.Err(err).Any("body", body).Msg("failed to marshal body")
			return
		}
		buf = bytes.NewBuffer(b)
	}

	var req *http.Request
	if req, err = http.NewRequest(method, url, buf); err != nil {
		l.Err(err).Msg("failed to create request")
		return
	}

	req.Header = map[string][]string{"x-api-key": {os.Getenv("WEB_KEY")}}
	if method == http.MethodPut {
		req.Header.Set("Content-Type", "application/json")
	}

	var res *http.Response
	if res, err = _client.Do(req); err != nil {
		l.Err(err).Msg("failed to do request")
		return
	}
	defer func() { _ = res.Body.Close() }()

	if b, err = io.ReadAll(res.Body); err != nil {
		l.Err(err).Msg("failed to read response")
	} else if res.StatusCode > 299 {
		err = fmt.Errorf("[%d] %s", res.StatusCode, string(b))
		l.Err(err).Msg("request failed")
	}
	return
}
