package client

import (
	"bytes"
	"context"

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
