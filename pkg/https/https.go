package https

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/rs/zerolog/log"
)

func Get(url string) ([]byte, error) {

	var errs error

	for i := 0; i < 3; i++ {

		data, code, err := get(url)
		if code < 300 {
			return data, nil
		}

		errs = errors.Join(errs, err)

		if code != 429 {
			break
		}

		time.Sleep(time.Second * time.Duration(5*(i+1)))
	}

	return nil, errs
}

func get(url string) ([]byte, int, error) {

	log.Trace().Str("url", url).Msg("get")

	res, err := http.Get(url)
	if err != nil {
		log.Err(err).Str("url", url).Msg("failed to get")
		return nil, -1, err
	}
	defer res.Body.Close()

	log.Debug().
		Str("url", url).
		Str("status", res.Status).
		Msg("got")

	var b []byte
	b, err = io.ReadAll(res.Body)

	return b, res.StatusCode, err
}

func PostForm(u string, v url.Values) ([]byte, error) {
	return post(u, []byte(v.Encode()), map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	})
}

func PostJSON(u string, b []byte, h map[string]string) ([]byte, error) {
	h["Content-Type"] = "application/json"
	return post(u, b, h)
}

func post(u string, b []byte, h map[string]string) ([]byte, error) {

	req, err := http.NewRequest(http.MethodPost, u, bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}

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
