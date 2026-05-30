package password

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreate(t *testing.T) {
	txt := "*mZ6EYi6Z3hcJ!"
	pid, err := Create(txt)
	assert.NoError(t, err)
	assert.NotEmpty(t, pid)
	t.Log(pid)

}
