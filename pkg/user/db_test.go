package user

import (
	"testing"

	"github.com/nelsw/bytelyon/pkg/id"
	"github.com/nelsw/bytelyon/pkg/logs"
	"github.com/stretchr/testify/assert"
)

func TestExists(t *testing.T) {
	logs.Init("debug")
	m, err := Find("01KM010XK0HY8HWWFPJTZGRF0F")
	assert.NoError(t, err)
	t.Log(m)
	uid, e := Find(id.ParseULID("01KM010XK0HY8HWWFPJTZGRF0F"))
	assert.NoError(t, e)
	t.Log(m)
	t.Log(uid)
}

func TestSave(t *testing.T) {

	m, err := Find("01KM010XK0HY8HWWFPJTZGRF0F")
	assert.NoError(t, err)
	m.PID = id.ParseUUID("019e7564-f2cd-74be-aa1f-81d7507d6a2e")
	m.EID = id.ParseUUID("978e639d-083b-5732-9586-8877097de56c")
	err = Save(m)
	assert.NoError(t, err)
}
