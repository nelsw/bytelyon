package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/ses"
)

var S3 *s3.Client
var SES *ses.Client

func init() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic("aws configuration error: " + err.Error())
	}
	S3 = s3.NewFromConfig(cfg)
	SES = ses.NewFromConfig(cfg)
}
