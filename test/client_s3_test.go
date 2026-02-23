package test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/nelsw/bytelyon/internal/client/fetch"
	"github.com/nelsw/bytelyon/internal/client/s3"
	"github.com/nelsw/bytelyon/internal/config"
	"github.com/nelsw/bytelyon/internal/db"
	"github.com/nelsw/bytelyon/internal/model"
	"github.com/stretchr/testify/assert"
)

func Test_Client_S3_New(t *testing.T) {

	ctx := context.Background()

	err := db.Builder[model.Settings]().Create(ctx, &model.Settings{
		AWS: model.AWS{
			Credentials: aws.Credentials{
				AccessKeyID:     config.Get[string]("AWS_ACCESS_KEY_ID"),
				SecretAccessKey: config.Get[string]("AWS_SECRET_ACCESS_KEY"),
				Source:          "StaticCredentialsProvider",
			},
			Bucket: "bytelyon",
			Region: "us-east-1",
		},
	})
	assert.NoError(t, err)

	var val model.Settings
	val, err = db.Builder[model.Settings]().First(ctx)

	var data = struct{ Hello string }{Hello: "World"}
	b, _ := json.Marshal(&data)

	client := s3.New(&val)

	err = client.Put("data.json", b)
	assert.NoError(t, err)

	b, err = client.Get("data.json")
	assert.NoError(t, err)
	assert.Equal(t, string(b), `{"Hello":"World"}`)

	var url string
	url, err = client.URL("data.json", 10)
	assert.NoError(t, err)
	assert.True(t, len(url) > 0)

	var fb []byte
	fb, err = fetch.New(url).Bytes()
	assert.NoError(t, err)
	assert.Equal(t, string(fb), `{"Hello":"World"}`)

	err = client.Delete("data.json")
	assert.NoError(t, err)
}
