package router

import (
	. "github.com/nelsw/bytelyon/pkg/api"
	"github.com/nelsw/bytelyon/pkg/handler/router/ai"
	"github.com/nelsw/bytelyon/pkg/handler/router/articles"
	"github.com/nelsw/bytelyon/pkg/handler/router/bots"
	"github.com/nelsw/bytelyon/pkg/handler/router/helper"
	"github.com/nelsw/bytelyon/pkg/handler/router/pages"
	"github.com/nelsw/bytelyon/pkg/handler/router/sitemaps"
)

func Handler(r Request) (Response, error) {

	r.Log()

	switch r.RawPath {
	case "/v1/ai":
		return ai.Handler(r), nil
	case "/v1/bots":
		return bots.Handler(r), nil
	case "/v1/articles":
		return articles.Handler(r), nil
	case "/v1/sitemaps":
		return sitemaps.Handler(r), nil
	case "/v1/pages":
		return pages.Handler(r), nil
	case "/v1/helper":
		return helper.Handler(r), nil
	}

	return r.NI(), nil
}
