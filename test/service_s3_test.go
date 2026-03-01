package test

import (
	"testing"
	"time"

	"github.com/nelsw/bytelyon/internal/client/fetch"
	"github.com/nelsw/bytelyon/internal/service/s3"
	"github.com/stretchr/testify/assert"
)

func Test_Service_S3(t *testing.T) {

	type Data struct {
		Hello string `json:"hello"`
	}

	assert.NoError(t, s3.Save("data.json", Data{Hello: "World"}))

	data, err := s3.Find[Data]("data.json")
	assert.NoError(t, err)
	assert.Equal(t, "World", data.Hello)

	var url string
	url, err = s3.Share("data.json", time.Minute)
	assert.NoError(t, err)
	assert.True(t, len(url) > 0)

	var fb []byte
	fb, err = fetch.New(url).Bytes()
	assert.NoError(t, err)
	assert.Equal(t, `{"Hello":"World"}`, string(fb))

	err = s3.Wipe("data.json")
	assert.NoError(t, err)
}
