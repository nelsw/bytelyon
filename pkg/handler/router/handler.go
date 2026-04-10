package router

import (
	. "github.com/nelsw/bytelyon/pkg/api"
	"github.com/nelsw/bytelyon/pkg/handler/router/articles"
	"github.com/nelsw/bytelyon/pkg/handler/router/bots"
)

func Handler(r Request) (Response, error) {

	r.Log()

	switch r.RawPath {
	case "/v1/bots":
		return bots.Handle(r), nil
	case "/v1/articles":
		return articles.Handle(r), nil
	}

	return r.NI(), nil
}
