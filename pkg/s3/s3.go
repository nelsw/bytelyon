package s3

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/nelsw/bytelyon/pkg/util/ptr"
	"github.com/rs/zerolog/log"
)

var c *s3.Client

func init() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic("aws configuration error, " + err.Error())
	}
	c = s3.NewFromConfig(cfg)
}

func PutPrivateObject(key string, data []byte) error {
	return Put(key, data, false)
}

func PutPublicImage(key string, data []byte) (string, error) {
	return "https://bytelyon-public.s3.amazonaws.com/" + key, Put(key, data, true)
}

// Put creates a new object or replaces an old object with a new object.
func Put(key string, data []byte, isPublic bool) error {

	var bucket string
	if isPublic {
		bucket = "bytelyon-public"
	} else {
		bucket = "bytelyon-private"
	}

	l := log.With().
		Str("ƒ", "put").
		Str("bucket", bucket).
		Str("key", key).
		Int("body", len(data)).
		Logger()

	l.Trace().Send()

	if len(key) == 0 {
		return errors.New("cannot put object with empty key")
	} else if len(data) == 0 {
		return errors.New("cannot put object with empty data")
	}

	in := &s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &key,
		Body:   bytes.NewReader(data),
	}

	if isPublic {
		in.ACL = types.ObjectCannedACLPublicRead
		in.ContentType = ptr.Of(http.DetectContentType(data))
	}

	if _, err := c.PutObject(context.Background(), in); err != nil {
		l.Err(err).Send()
		return err
	}

	l.Debug().Send()
	return nil
}

func GetPrivateObject(key string) ([]byte, error) {

	bucket := "bytelyon-private"

	l := log.With().
		Str("ƒ", "GetPrivateObject").
		Str("bucket", bucket).
		Str("key", key).
		Logger()

	l.Trace().Send()

	out, err := c.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})

	if err != nil {
		l.Err(err).Send()
		return nil, err
	}

	var b []byte
	if b, err = io.ReadAll(out.Body); err != nil {
		l.Err(err).Send()
		return nil, err
	}

	if err = out.Body.Close(); err != nil {
		l.Err(err).Send()
	}

	l.Debug().Send()

	return b, nil
}

func Get(key string, isPublic bool) ([]byte, error) {

	var bucket string
	if isPublic {
		bucket = "bytelyon-public"
	} else {
		bucket = "bytelyon-private"
	}

	l := log.With().
		Str("ƒ", "GetPrivateObject").
		Str("bucket", bucket).
		Str("key", key).
		Logger()

	l.Trace().Send()

	out, err := c.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})

	if err != nil {
		l.Err(err).Send()
		return nil, err
	}

	var b []byte
	if b, err = io.ReadAll(out.Body); err != nil {
		l.Err(err).Send()
		return nil, err
	}

	if err = out.Body.Close(); err != nil {
		l.Err(err).Send()
	}

	l.Debug().Send()

	return b, nil
}

func DeletePrivateObject(key string) error {

	bucket := "bytelyon-private"

	l := log.With().
		Str("ƒ", "delete").
		Str("bucket", bucket).
		Str("key", key).
		Logger()

	l.Trace().Send()

	_, err := c.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})

	if err != nil {
		l.Err(err).Send()
		return err
	}

	l.Debug().Send()
	return nil
}

func DeletePublicImage(key string) error {
	bucket := "bytelyon-public"

	l := log.With().
		Str("ƒ", "delete").
		Str("bucket", bucket).
		Str("key", key).
		Logger()

	l.Trace().Send()

	_, err := c.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})

	if err != nil {
		l.Err(err).Send()
		return err
	}

	l.Debug().Send()
	return nil
}

func Delete(key string, isPublic bool) error {
	var bucket string
	if isPublic {
		bucket = "bytelyon-public"
	} else {
		bucket = "bytelyon-private"
	}

	l := log.With().
		Str("ƒ", "delete").
		Str("bucket", bucket).
		Str("key", key).
		Logger()

	l.Trace().Send()

	_, err := c.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})

	if err != nil {
		l.Err(err).Send()
		return err
	}

	l.Debug().Send()
	return nil
}

func ListDirectories(prefix string, after ...string) ([]string, error) {

	var startAfter string
	if len(after) > 0 {
		startAfter = after[0]
	}

	bucket := "bytelyon-private"

	l := log.With().
		Str("ƒ", "ListDirectories").
		Str("bucket", bucket).
		Str("prefix", prefix).
		Str("startAfter", startAfter).
		Logger()

	l.Trace().Send()

	out, err := c.ListObjectsV2(context.Background(), &s3.ListObjectsV2Input{
		Bucket:     &bucket,
		Prefix:     &prefix,
		StartAfter: &startAfter,
	})
	if err != nil {
		l.Err(err).Send()
		return nil, err
	}

	var keys []string
	for _, obj := range out.Contents {
		keys = append(keys, *obj.Key)
	}

	return keys, nil
}

func GetPresignedURL(key string) (string, error) {
	client := s3.NewPresignClient(c)

	presignedUrl, err := client.PresignGetObject(context.Background(), &s3.GetObjectInput{
		Bucket: ptr.Of("bytelyon-public"),
		Key:    &key,
	}, s3.WithPresignExpires(30*time.Minute))

	if err != nil {
		return "", err
	}

	return presignedUrl.URL, nil
}
