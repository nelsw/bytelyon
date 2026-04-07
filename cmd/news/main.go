package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/nelsw/bytelyon/pkg/handler/news"
)

func main() { lambda.Start(news.Handler) }
