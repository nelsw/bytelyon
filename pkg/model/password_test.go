package model

import (
	"testing"

	"github.com/nelsw/bytelyon/internal/logger"
	"github.com/nelsw/bytelyon/pkg/db"
	"github.com/stretchr/testify/assert"
)

func TestPassword(t *testing.T) {
	logger.Init()

	userID := NewULID()
	pass := Password{UserID: userID}
	text := "publixChickenTenderSubsOnSale!!5.99"

	err := pass.Generate(text)
	assert.NoError(t, err)
	assert.NotEmpty(t, pass.Hash)
	assert.NoError(t, pass.Compare(text))

	err = db.Put(pass)
	assert.NoError(t, err)

	var p Password

	p, err = db.Get[Password](&Password{UserID: userID})
	assert.NoError(t, err)
	assert.Equal(t, pass.UserID, p.UserID)

	err = db.Delete(p)
	assert.NoError(t, err)

	p, err = db.Get[Password](&Password{UserID: userID})
	assert.Empty(t, p)
}
