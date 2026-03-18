package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/nelsw/bytelyon/pkg/handler/health"
)

func main() { lambda.Start(health.Handler) }
