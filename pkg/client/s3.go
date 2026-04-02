package client

import (
	"bytes"
	"context"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/rs/zerolog/log"
)

// PutPublicImage creates a new object or replaces an old object with a new object.
func PutPublicImage(ctx context.Context, c *s3.Client, bucket, key string, b []byte) error {
	l := log.With().
		Str("ƒ", "PutPublicImage").
		Str("bucket", bucket).
		Str("key", key).
		Int("body", len(b)).
		Logger()

	l.Info().Send()
	_, err := c.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &bucket,
		Key:         &key,
		Body:        bytes.NewReader(b),
		ACL:         types.ObjectCannedACLPublicRead,
		ContentType: util.Ptr(http.DetectContentType(b)),
	})
	l.Err(err).Send()
	return err
}

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
