package test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/nelsw/bytelyon/internal/client/fetch"
	"github.com/nelsw/bytelyon/internal/client/s3"
	"github.com/nelsw/bytelyon/internal/config"
	"github.com/stretchr/testify/assert"
)

func Test_Client_S3(t *testing.T) {

	var err error
	ctx := context.Background()
	c := client.New(
		config.Get[string]("AWS_REGION"),
		config.Get[string]("AWS_ACCESS_KEY_ID"),
		config.Get[string]("AWS_SECRET_ACCESS_KEY"),
	)
	bucket := config.Get[string]("AWS_S3_BUCKET")

	var data = struct{ Hello string }{Hello: "World"}
	b, _ := json.Marshal(&data)

	err = client.PutObject(ctx, c, bucket, "data.json", b)
	assert.NoError(t, err)

	b, err = client.GetObject(ctx, c, bucket, "data.json")
	assert.NoError(t, err)
	assert.Equal(t, `{"Hello":"World"}`, string(b))

	var url string
	url, err = client.PresignGetObject(ctx, c, bucket, "data.json", time.Minute)
	assert.NoError(t, err)
	assert.True(t, len(url) > 0)

	var fb []byte
	fb, err = fetch.New(url).Bytes()
	assert.NoError(t, err)
	assert.Equal(t, `{"Hello":"World"}`, string(fb))

	err = client.DeleteObject(ctx, c, bucket, "data.json")
	assert.NoError(t, err)
}
