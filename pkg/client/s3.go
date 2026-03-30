package client

import (
	"bytes"
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/rs/zerolog/log"
)

// PutObject creates a new object or replaces an old object with a new object.
func PutObject(ctx context.Context, c *s3.Client, bucket, key string, bdy []byte) error {
	l := log.With().
		Str("bucket", bucket).
		Str("key", key).
		Int("body", len(bdy)).
		Logger()

	l.Info().Msg("putting object")

	_, err := c.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &key,
		Body:   bytes.NewReader(bdy),
	})

	if err != nil {
		l.Err(err).Msg("failed to put object")
		return err
	}

	l.Info().Msg("put object")

	return nil
}

// GetObject retrieves an object from S3.
func GetObject(ctx context.Context, c *s3.Client, bucket, key string) ([]byte, error) {

	l := log.With().
		Str("bucket", bucket).
		Str("key", key).
		Logger()

	l.Info().Msg("getting object")

	resp, err := c.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})

	if err != nil {
		l.Err(err).Msg("failed to get object")
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	var b []byte
	if b, err = io.ReadAll(resp.Body); err != nil {
		l.Err(err).Msg("failed to read object")
	}
	return b, err
}
