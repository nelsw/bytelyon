package s3

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/nelsw/bytelyon/pkg/aws"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/rs/zerolog/log"
)

func PutPrivateObject(key string, data []byte) error {
	return put(key, data, false)
}

func PutPublicImage(key string, data []byte) (string, error) {
	return "https://bytelyon-public.s3.amazonaws.com/" + key, put(key, data, true)
}

func PutPrivateBotData(b *model.Bot, k string, d []byte) (string, error) {
	return putBotData(b, k, d, false)
}

func PutPublicBotData(b *model.Bot, k string, d []byte) (string, error) {
	return putBotData(b, k, d, true)
}

func putBotData(b *model.Bot, k string, d []byte, isPublic bool) (string, error) {

	k = fmt.Sprintf(
		"users/%s/bots/%s/%s/%s",
		b.UserID,
		b.Type,
		strings.ReplaceAll(b.Target, " ", "-"),
		k,
	)

	if isPublic {
		return PutPublicImage(k, d)
	}
	return k, PutPrivateObject(k, d)
}

// put creates a new object or replaces an old object with a new object.
func put(key string, data []byte, isPublic bool) error {

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
		in.ContentType = util.Ptr(http.DetectContentType(data))
	}

	if _, err := aws.S3().PutObject(context.Background(), in); err != nil {
		l.Err(err).Send()
		return err
	}

	l.Debug().Send()
	return nil
}

func GetPrivateObject(key string) ([]byte, error) {

	bucket := "bytelyon-private"

	l := log.With().
		Str("ƒ", "put").
		Str("bucket", bucket).
		Str("key", key).
		Bool("isPublic", false).
		Logger()

	l.Trace().Send()

	out, err := aws.S3().GetObject(context.Background(), &s3.GetObjectInput{
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
	out.Body.Close()

	l.Debug().Send()

	return b, nil
}
