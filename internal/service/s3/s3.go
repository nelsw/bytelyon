package s3

import (
	"context"
	"encoding/json"
	"time"

	aws "github.com/aws/aws-sdk-go-v2/config"
	S3 "github.com/aws/aws-sdk-go-v2/service/s3"
	client "github.com/nelsw/bytelyon/internal/client/s3"
	"github.com/nelsw/bytelyon/internal/config"
	"github.com/nelsw/bytelyon/internal/util"
)

var (
	s3  *S3.Client
	ctx context.Context

	bucket = func() string { return "bytelyon-db-" + config.Mode() }
)

func init() {
	ctx = context.Background()
	c := util.Must(aws.LoadDefaultConfig(ctx))
	s3 = S3.NewFromConfig(c)
}

// Wipe removes an object from the s3 bucket.
func Wipe(key string) error {
	return client.DeleteObject(ctx, s3, bucket(), key)
}

// Find gets an object from a s3 bucket and marshals it into the given type.
func Find[T any](key string) (t T, err error) {
	var b []byte
	if b, err = client.GetObject(ctx, s3, bucket(), key); err != nil {
		return
	}
	err = json.Unmarshal(b, &t)
	return
}

func Save(key string, b []byte) (err error) {
	return client.PutObject(ctx, s3, bucket(), key, b)
}

func Share(key string, exp time.Duration) (string, error) {
	return client.PresignGetObject(ctx, s3, bucket(), key, exp)
}
