package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/nelsw/bytelyon/pkg/handler/bots/news"
)

func main() { lambda.Start(bots.Handler) }
