package news

import (
	"github.com/nelsw/bytelyon/pkg/api"
)

func Handler(r api.Request) api.Response {

	return r.NI()
}
