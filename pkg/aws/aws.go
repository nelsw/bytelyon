package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

var cfg aws.Config

func init() {
	var err error
	if cfg, err = config.LoadDefaultConfig(context.Background()); err != nil {
		panic(err)
	}
}

func DB() *dynamodb.Client {
	return dynamodb.NewFromConfig(cfg)
}
