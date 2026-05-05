package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	cfg aws.Config
	dbc *dynamodb.Client
	s3c *s3.Client
)

func Init(args ...string) {
	if len(args) == 0 {
		cfg, _ = config.LoadDefaultConfig(context.Background())
	} else {
		cfg = aws.Config{
			Credentials: credentials.NewStaticCredentialsProvider(args[0], args[1], ""),
			Region:      args[2],
		}
	}
}

func DB() *dynamodb.Client {
	if dbc == nil {
		Init()
		dbc = dynamodb.NewFromConfig(cfg)
	}
	return dbc
}

func S3() *s3.Client {
	if s3c == nil {
		Init()
		s3c = s3.NewFromConfig(cfg)
	}
	return s3c
}
