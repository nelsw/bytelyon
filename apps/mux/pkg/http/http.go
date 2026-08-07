package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

type Header = http.Header

var client = http.Client{
	Timeout: 10 * time.Second,
}

func Get[T any](url string, header Header) (t T, err error) {
	var out []byte
	if out, err = do(http.MethodGet, url, nil, header); err != nil {
		log.Err(err).Msgf("failed to get %s", url)
	} else if err = json.Unmarshal(out, &t); err != nil {
		log.Err(err).Msgf("failed to unmarshal response %s", string(out))
	}
	return
}

func Put(url string, data any, header Header) (b []byte, err error) {
	return do(http.MethodPut, url, data, header)
}

func do(method, url string, body any, header Header) (b []byte, err error) {

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

	if header != nil {
		req.Header = header
	}
	if method == http.MethodPut {
		req.Header.Set("Content-Type", "application/json")
	}

	var res *http.Response
	if res, err = client.Do(req); err != nil {
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
