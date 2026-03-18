package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/nelsw/bytelyon/pkg/handler/bot"
)

func main() { lambda.Start(bot.Handler) }
