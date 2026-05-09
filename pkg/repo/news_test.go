package repo

import (
	"encoding/json"
	"testing"

	"github.com/nelsw/bytelyon/pkg/logs"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"
)

func TestGetNews(t *testing.T) {
	logs.Init("info")
	out, err := GetNews(ulid.MustParse("01KM010XK0HY8HWWFPJTZGRF0F"), ulid.MustParse("01KMXJS68EKK50P412N58HSPSA"))
	assert.NoError(t, err)
	assert.NotEmpty(t, out)
	b, _ := json.MarshalIndent(out, "", "\t")
	t.Log(string(b))
}
