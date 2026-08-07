package s3

import (
	"bytes"
	"context"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/rs/zerolog/log"
)

var _s3 *s3.Client

func init() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic("aws configuration error, " + err.Error())
	}
	_s3 = s3.NewFromConfig(cfg)
}

func Save(key string, data []byte) {

	l := log.With().Str("key", key).Logger()

	if len(data) == 0 {
		l.Warn().Send()
		return
	}

	_, err := _s3.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:      aws.String("bytelyon-private"),
		Key:         &key,
		Body:        bytes.NewReader(data),
		ContentType: aws.String(http.DetectContentType(data)),
	})

	if err != nil {
		l.Err(err).Send()
		return
	}

	l.Trace().Send()
}
