package password

import (
	"testing"

	"github.com/nelsw/bytelyon/pkg/logs"
	"github.com/stretchr/testify/assert"
)

func TestCreate(t *testing.T) {
	logs.Init("debug")
	txt := "Demo123!"
	pid, err := Create(txt)
	assert.NoError(t, err)
	assert.NotEmpty(t, pid)
	t.Log(pid)

}
