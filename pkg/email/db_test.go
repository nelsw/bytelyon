package email

import (
	"testing"

	"github.com/nelsw/bytelyon/pkg/id"
	"github.com/nelsw/bytelyon/pkg/logs"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"
)

func TestCreate(t *testing.T) {
	logs.Init("debug")
	uid := id.ParseULID("01KM01JC9PS1R4X4FDJNFAR4AZ")
	txt := "demo@demo.com"
	err := Create(uid, txt)
	assert.NoError(t, err)
}

func TestFind(t *testing.T) {
	logs.Init("debug")
	uid, err := Find("demo@demo.com")
	assert.NoError(t, err)
	assert.NotEmpty(t, uid)
	assert.NotEqual(t, uid, ulid.Zero)
}
