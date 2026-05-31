package email

import (
	"testing"

	"github.com/nelsw/bytelyon/pkg/id"
	"github.com/nelsw/bytelyon/pkg/logs"
	"github.com/stretchr/testify/assert"
)

func TestCreate(t *testing.T) {
	logs.Init("debug")
	uid := id.ParseULID("01KM010XK0HY8HWWFPJTZGRF0F")
	txt := "carl@firefibers.com"
	err := Create(uid, txt)
	assert.NoError(t, err)
}
