package s3

import (
	"bytes"
	"context"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/nelsw/bytelyon/apps/mux/internal/util/gzip"
	"github.com/rs/zerolog/log"
)

var s3c *s3.Client

func Put(key string, body []byte, compressed bool) {

	var err error
	if compressed {
		if body, err = gzip.Decompress(body); err != nil {
			return
		}
	}

	_, err = s3c.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:      aws.String(os.Getenv("S3_BUCKET")),
		Key:         aws.String(key),
		Body:        bytes.NewReader(body),
		ContentType: aws.String(http.DetectContentType(body)),
	})

	if err != nil {
		log.Err(err).
			Str("bucket", os.Getenv("S3_BUCKET")).
			Str("key", key).
			Msg("failed to put object")
	}
}
