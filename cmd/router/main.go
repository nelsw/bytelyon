package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/nelsw/bytelyon/pkg/api"
	"github.com/nelsw/bytelyon/pkg/handler/router/ai"
	"github.com/nelsw/bytelyon/pkg/handler/router/bots"
	"github.com/nelsw/bytelyon/pkg/handler/router/helper"
	"github.com/nelsw/bytelyon/pkg/handler/router/news"
	"github.com/nelsw/bytelyon/pkg/handler/router/pages"
	"github.com/nelsw/bytelyon/pkg/handler/router/searches"
	"github.com/nelsw/bytelyon/pkg/handler/router/shopify"
	"github.com/nelsw/bytelyon/pkg/handler/router/sitemaps"
	"github.com/nelsw/bytelyon/pkg/sitemap"
)

func Handler(r api.Request) (api.Response, error) {

	r.Log()

	switch r.RawPath {
	case "/v1/ai":
		return ai.Handler(r), nil
	case "/v1/bots":
		return bots.Handler(r), nil
	case "/v1/helper":
		return helper.Handler(r), nil

	case "/v1/shopify":
		return shopify.Handler(r), nil

	case "/v1/sitemap":
		return sitemap.Handler(r), nil

	/*
		deprecated
	*/
	case "/v1/news":
		return news.Handler(r), nil
	case "/v1/page":
		return pages.Handler(r), nil
	case "/v1/searches":
		return searches.Handler(r), nil
	case "/v1/sitemaps":
		return sitemaps.Handler(r), nil
	}

	return r.NI(), nil
}

func main() { lambda.Start(Handler) }
