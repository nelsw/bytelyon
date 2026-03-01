package store

import (
	"context"
	"encoding/json"
	"time"

	aws "github.com/aws/aws-sdk-go-v2/config"
	S3 "github.com/aws/aws-sdk-go-v2/service/s3"
	client "github.com/nelsw/bytelyon/internal/client/s3"
)

var (
	s3  *S3.Client
	ctx context.Context
)

func init() {
	ctx = context.Background()
	c, err := aws.LoadDefaultConfig(ctx)
	if err != nil {
		panic(err)
	}
	s3 = S3.NewFromConfig(c)
}

// Wipe removes an object from a s3 bucket.
func Wipe(key string) error {
	return client.DeleteObject(ctx, s3, key)
}

// Find gets an object from a s3 bucket and marshals it into the given type.
func Find[T any](key string) (t T, err error) {
	var b []byte
	if b, err = client.GetObject(ctx, s3, key); err != nil {
		return
	}
	err = json.Unmarshal(b, &t)
	return
}

func Save(key string, a any) (err error) {
	var b []byte
	switch t := a.(type) {
	case []byte:
		break
	case string:
		b = []byte(t)
	default:
		if b, err = json.Marshal(a); err != nil {
			return err
		}
	}
	return client.PutObject(ctx, s3, key, b)
}

func Share(key string, exp time.Duration) (string, error) {
	return client.PresignGetObject(ctx, s3, key, exp)
}
