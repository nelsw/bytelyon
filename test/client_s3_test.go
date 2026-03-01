package test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/nelsw/bytelyon/internal/client/fetch"
	"github.com/nelsw/bytelyon/internal/client/s3"
	"github.com/nelsw/bytelyon/internal/util"
	"github.com/stretchr/testify/assert"
)

func Test_Client_S3(t *testing.T) {

	var err error
	ctx := context.Background()
	c := util.Must(client.New())

	var data = struct{ Hello string }{Hello: "World"}
	b, _ := json.Marshal(&data)

	err = client.PutObject(ctx, c, "data.json", b)
	assert.NoError(t, err)

	b, err = client.GetObject(ctx, c, "data.json")
	assert.NoError(t, err)
	assert.Equal(t, `{"Hello":"World"}`, string(b))

	var url string
	url, err = client.PresignGetObject(ctx, c, "data.json", time.Minute)
	assert.NoError(t, err)
	assert.True(t, len(url) > 0)

	var fb []byte
	fb, err = fetch.New(url).Bytes()
	assert.NoError(t, err)
	assert.Equal(t, `{"Hello":"World"}`, string(fb))

	err = client.DeleteObject(ctx, c, "data.json")
	assert.NoError(t, err)
}
