package https

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/nelsw/bytelyon/pkg/util/urls"
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

func PostForm(u string, v url.Values) ([]byte, error) {
	return post(u, []byte(v.Encode()), map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	})
}

func PostJSON(u string, b []byte, h map[string]string) ([]byte, error) {
	h["Content-Type"] = "application/json"
	return post(u, b, h)
}

func get(u string) ([]byte, int, error) {

	l := log.With().
		Str("ƒ", "get").
		Str("domain", urls.Domain(u)).
		Str("path", urls.Path(u)).
		Str("query", urls.Query(u)).
		Logger()

	l.Trace().Send()

	res, err := http.Get(u)
	if err != nil {
		l.Err(err).Send()
		return nil, -1, err
	}
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			l.Err(closeErr).Send()
		}
	}()

	l.Trace().
		Str("status", res.Status).
		Send()

	var b []byte
	b, err = io.ReadAll(res.Body)

	return b, res.StatusCode, err
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
