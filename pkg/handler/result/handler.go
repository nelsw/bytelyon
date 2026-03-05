package result

import (
	. "github.com/nelsw/bytelyon/pkg/api"
)

func Handler(r Request) (Response, error) {

	r.Log()

	return r.NI(), nil
}

func DeleteResult(r Request) Response {
	return r.NC()
}

func GetResults(r Request) Response {

}
