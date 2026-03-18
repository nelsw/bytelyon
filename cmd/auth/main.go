package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/nelsw/bytelyon/pkg/handler/auth"
)

func main() { lambda.Start(auth.Handler) }
