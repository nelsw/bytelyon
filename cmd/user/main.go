package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/nelsw/bytelyon/pkg/handler/user"
)

func main() { lambda.Start(user.Handler) }
