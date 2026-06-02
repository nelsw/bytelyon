package s3

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"

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
		Str("key", key).
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

func Get(key string, isPublic bool) ([]byte, error) {

	var bucket string
	if isPublic {
		bucket = "bytelyon-public"
	} else {
		bucket = "bytelyon-private"
	}

	l := log.With().
		Str("ƒ", "get").
		Str("key", key).
		Logger()

	l.Trace().Send()

	out, err := c.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})

	if err != nil {
		if !strings.Contains(err.Error(), "NoSuchKey") {
			l.Err(err).Send()
		}
		return nil, err
	}

	var b []byte
	if b, err = io.ReadAll(out.Body); err != nil {
		l.Warn().Err(err).Send()
		return nil, err
	}

	if err = out.Body.Close(); err != nil {
		l.Warn().Err(err).Send()
	}

	l.Debug().Send()

	return b, nil
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
		Str("key", key).
		Logger()

	l.Trace().Send()

	_, err := c.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})

	if err != nil {
		l.Warn().Err(err).Send()
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
