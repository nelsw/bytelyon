package helper

import (
	"encoding/json"
	"errors"

	. "github.com/nelsw/bytelyon/pkg/api"
)

func Handler(r Request) Response {

	r.Log()

	if r.IsGuest() {
		return r.NOPE()
	}

	if r.Query("err") != "" {
		return r.BAD(errors.New("test error"))
	}

	var m map[string]any
	if err := json.Unmarshal([]byte(r.Body), &m); err != nil {
		return r.BAD(err)
	}

	return r.OK(m)
}
