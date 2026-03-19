package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/nelsw/bytelyon/pkg/handler/authorizer"
)

func main() { lambda.Start(authorizer.Handler) }
