package model

import (
	"testing"

	"github.com/nelsw/bytelyon/pkg/db"
	"github.com/stretchr/testify/assert"
)

func TestPassword(t *testing.T) {
	t.Setenv("MODE", "release")
	//db.Migrate(&Password{})
	userID := NewULID()
	pass := &Password{UserID: userID}
	text := "Demo123!"

	err := pass.Generate(text)
	assert.NoError(t, err)
	assert.NotEmpty(t, pass.Hash)
	assert.NoError(t, pass.Compare(text))

	err = db.Put(pass)
	assert.NoError(t, err)

	var p *Password

	p, err = db.Get(&Password{UserID: userID})
	assert.NoError(t, err)
	assert.Equal(t, pass.UserID, p.UserID)

	//err = db.Delete(p)
	//assert.NoError(t, err)

}
