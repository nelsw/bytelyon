package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/nelsw/bytelyon/pkg/handler/router"
)

func main() { lambda.Start(router.Handler) }
