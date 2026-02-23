package db

import (
	"bytes"
	"context"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	_s3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/nelsw/bytelyon/internal/util"
	"github.com/rs/zerolog/log"
)

type S3 interface {
	Delete(context.Context, string, string) error
	Get(context.Context, string, string) ([]byte, error)
	Put(context.Context, string, string, []byte) error
	URL(context.Context, string, string, int64) (string, error)
}

type s3 struct {
	*_s3.Client
	*_s3.PresignClient
}

func (s3 *s3) Delete(ctx context.Context, bucket, key string) error {

	_, err := s3.DeleteObject(ctx, &_s3.DeleteObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})

	log.Debug().
		Err(err).
		Str("bucket", bucket).
		Str("key", key).
		Msg("Delete")

	return err
}

func (s3 *s3) Get(ctx context.Context, bucket, key string) ([]byte, error) {

	out, err := s3.GetObject(ctx, &_s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		log.Warn().Err(err).Msg("Get - Failed to read body")
		return nil, err
	}
	var body []byte
	defer out.Body.Close()
	if body, err = io.ReadAll(out.Body); err != nil {
		log.Warn().Err(err).Msg("Get - Failed to read body")
		return nil, err
	}

	log.Trace().
		Err(err).
		Str("bucket", bucket).
		Str("key", key).
		Bytes("body", body).
		Msg("Get")

	return body, err
}

func (s3 *s3) Put(ctx context.Context, bucket, key string, b []byte) (err error) {

	_, err = s3.PutObject(ctx, &_s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &key,
		Body:   bytes.NewReader(b),
	})

	log.Trace().
		Err(err).
		Str("bucket", bucket).
		Str("key", key).
		Bytes("body", b).
		Msg("Put")

	return
}

func (s3 *s3) URL(ctx context.Context, bucket, key string, i int64) (string, error) {

	out, err := s3.PresignGetObject(ctx, &_s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}, _s3.WithPresignExpires(time.Duration(i)*time.Hour))

	var url string
	if err == nil && out != nil {
		url = out.URL
	}

	log.Trace().
		Err(err).
		Str("bucket", bucket).
		Str("key", key).
		Int64("exp", i).
		Str("url", url).
		Msg("URL")

	return url, err
}

// NewS3 returns a new S3 client with a background context.
// An optional variadic set of Config values can be provided as
// input that will be prepended to the configs slice.
func NewS3(optFns ...func(*config.LoadOptions) error) S3 {
	return NewS3WithContext(context.Background(), optFns...)
}

// NewS3WithContext returns a new S3 client with the provided context.
// An optional variadic set of Config values can be provided as
// input that will be prepended to the configs slice.
func NewS3WithContext(ctx context.Context, optFns ...func(*config.LoadOptions) error) S3 {
	cfg := util.Must(config.LoadDefaultConfig(ctx, optFns...))
	cc := _s3.NewFromConfig(cfg)
	pc := _s3.NewPresignClient(cc)
	return &s3{cc, pc}
}
