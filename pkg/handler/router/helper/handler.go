package helper

import (
	"encoding/json"
	"errors"

	. "github.com/nelsw/bytelyon/pkg/api"
	"github.com/nelsw/bytelyon/pkg/model"
)

func Handler(r Request) Response {

	r.Log()

	if r.Query("err") != "" {
		return r.BAD(errors.New("test error"))
	}

	var m model.Data[any]
	if err := json.Unmarshal([]byte(r.Body), &m); err != nil {
		return r.BAD(err)
	}

	return r.OK(m)
}
