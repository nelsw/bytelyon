package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/nelsw/bytelyon/pkg/handler/bots/search/results"
)

func main() { lambda.Start(bots.Handler) }
