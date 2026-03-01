package client

import (
	"bytes"
	"context"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/nelsw/bytelyon/internal/config"
	"github.com/rs/zerolog/log"
)

var bucket = "bytelyon-db-" + config.Mode()

// DeleteObject removes an object from the s3 bucket with the given key.
func DeleteObject(ctx context.Context, c *s3.Client, key string) error {

	l := log.With().
		Str("bucket", bucket).
		Str("key", key).
		Logger()

	l.Trace().Msg("deleting object")

	_, err := c.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})

	if err != nil {
		l.Err(err).Msg("failed to delete object")
		return err
	}

	l.Debug().Msg("object deleted")

	return nil
}

// GetObject retrieves an object from the s3 bucket with the given key.
func GetObject(ctx context.Context, c *s3.Client, key string) ([]byte, error) {
	l := log.With().
		Str("bucket", bucket).
		Str("key", key).
		Logger()

	l.Trace().Msg("getting object")

	out, err := c.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})

	if err != nil {
		l.Err(err).Msg("failed to get object")
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		if err = Body.Close(); err != nil {
			l.Err(err).Msg("failed to close object body")
		}
	}(out.Body)

	var body []byte
	if body, err = io.ReadAll(out.Body); err != nil {
		l.Err(err).Msg("failed to read object body")
		return nil, err
	}

	l.Debug().Bytes("body", body).Msg("got object")

	return body, nil
}

// PutObject creates a new object, or replaces an old object with a new object.
func PutObject(ctx context.Context, c *s3.Client, key string, bdy []byte) error {
	l := log.With().
		Str("bucket", bucket).
		Str("key", key).
		Bytes("body", bdy).
		Logger()

	l.Trace().Msg("putting object")

	_, err := c.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &key,
		Body:   bytes.NewReader(bdy),
	})

	if err != nil {
		l.Err(err).Msg("failed to put object")
		return err
	}

	l.Debug().Msg("put object")

	return nil
}

// PresignGetObject returns a presigned HTTP Request which contains presigned URL, signed headers and HTTP method used.
func PresignGetObject(ctx context.Context, c *s3.Client, key string, exp time.Duration) (string, error) {

	l := log.With().
		Str("bucket", bucket).
		Str("key", key).
		Dur("exp", exp).
		Logger()

	out, err := s3.NewPresignClient(c).PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}, s3.WithPresignExpires(exp))

	if err != nil {
		l.Err(err).Msg("failed to presign object")
		return "", err
	}

	l.Debug().Str("url", out.URL).Msg("got presigned object")

	return out.URL, nil
}

// New returns a new s3.Client with the given Region, AccessKeyID, and SecretAccessKey.
func New(args ...context.Context) (*s3.Client, error) {
	return s3.NewFromConfig(config.Aws()), nil
}
