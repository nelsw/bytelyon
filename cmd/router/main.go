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
	defer res.Log()

	switch req.RawPath {
	case "/v1/ai":
		return ai.Handler(req), nil
	case "/v1/bots":
		return bot.Handler(req), nil
	case "/v1/shopify":
		return shopify.Handler(req), nil
	case "/v1/news":
		return news.Handler(req), nil
	case "/v1/searches":
		return search.Handler(req), nil
	case "/v1/sitemaps":
		return sitemap.Handler(req), nil
	}

	return api.NotImplemented(), nil
}

func main() { lambda.Start(Handler) }
