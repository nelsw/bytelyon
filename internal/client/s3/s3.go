package s3

import (
	"bytes"
	"context"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	_s3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/nelsw/bytelyon/internal/model"
	"github.com/nelsw/bytelyon/internal/util"
	"github.com/rs/zerolog/log"
)

type S3 interface {
	Delete(string) error
	Get(string) ([]byte, error)
	Put(string, []byte) error
	URL(string, int64) (string, error)
}

type s3 struct {
	bucket string
	context.Context
	*_s3.Client
	*_s3.PresignClient
}

func (s3 *s3) Delete(key string) error {

	_, err := s3.DeleteObject(s3.Context, &_s3.DeleteObjectInput{
		Bucket: &s3.bucket,
		Key:    &key,
	})

	log.Debug().
		Err(err).
		Str("bucket", s3.bucket).
		Str("key", key).
		Msg("Delete")

	return err
}

func (s3 *s3) Get(key string) ([]byte, error) {

	l := log.With().
		Str("bucket", s3.bucket).
		Str("key", key).
		Logger()

	out, err := s3.GetObject(s3.Context, &_s3.GetObjectInput{
		Bucket: &s3.bucket,
		Key:    &key,
	})

	if err != nil {
		l.Warn().Err(err).Msg("Get - Failed to read body")
		return nil, err
	}
	defer out.Body.Close()

	var body []byte
	if body, err = io.ReadAll(out.Body); err != nil {
		l.Warn().Err(err).Msg("Get - Failed to read body")
		return nil, err
	}

	l.Debug().Err(err).Bytes("body", body).Msg("Get")

	return body, err
}

func (s3 *s3) Put(key string, b []byte) (err error) {

	_, err = s3.PutObject(s3.Context, &_s3.PutObjectInput{
		Bucket: &s3.bucket,
		Key:    &key,
		Body:   bytes.NewReader(b),
	})

	log.Debug().
		Err(err).
		Str("bucket", s3.bucket).
		Str("key", key).
		Bytes("body", b).
		Msg("Put")

	return
}

func (s3 *s3) URL(key string, i int64) (string, error) {

	out, err := s3.PresignGetObject(s3.Context, &_s3.GetObjectInput{
		Bucket: &s3.bucket,
		Key:    &key,
	}, _s3.WithPresignExpires(time.Duration(i)*time.Hour))

	var url string
	if err == nil && out != nil {
		url = out.URL
	}

	log.Debug().
		Err(err).
		Str("bucket", s3.bucket).
		Str("key", key).
		Int64("exp", i).
		Str("url", url).
		Msg("URL")

	return url, err
}

// New returns a new S3 client with a background context.
func New(settings *model.Settings) S3 {
	return NewWithContext(context.Background(), settings)
}

// NewWithContext returns a new S3 client with the provided context.
func NewWithContext(ctx context.Context, settings *model.Settings) S3 {
	cfg := util.Must(config.LoadDefaultConfig(ctx, func(options *config.LoadOptions) error {
		options.Credentials = settings
		options.Region = settings.Region
		return nil
	}))
	cc := _s3.NewFromConfig(cfg)
	pc := _s3.NewPresignClient(cc)
	return &s3{settings.Bucket, ctx, cc, pc}
}
