package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/nelsw/bytelyon/pkg/ai"
	"github.com/nelsw/bytelyon/pkg/api"
	"github.com/nelsw/bytelyon/pkg/bot"
	"github.com/nelsw/bytelyon/pkg/news"
	"github.com/nelsw/bytelyon/pkg/search"
	"github.com/nelsw/bytelyon/pkg/shopify"
	"github.com/nelsw/bytelyon/pkg/sitemap"
)

func Handler(req api.HTTPRequest) (res api.HTTPResponse, err error) {

	req.Log()

	switch req.RawPath {
	case "/v1/ai":
		res = ai.Handler(req)
	case "/v1/bots":
		res = bot.Handler(req)
	case "/v1/shopify":
		res = shopify.Handler(req)
	case "/v1/news":
		res = news.Handler(req)
	case "/v1/searches":
		res = search.Handler(req)
	case "/v1/sitemaps":
		res = sitemap.Handler(req)
	default:
		res = api.NotImplemented()
	}

	res.Log()

	return
}

func main() { lambda.Start(Handler) }
