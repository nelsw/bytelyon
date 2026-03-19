package model

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/nelsw/bytelyon/pkg/db"
	"github.com/nelsw/bytelyon/pkg/logger"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"
)

func TestCredentials_Authenticate(t *testing.T) {

	logger.Init()

	//db.Migrate(Email{}, Password{}, User{})

	var err error

	email := &Email{
		Address: gofakeit.Email(),
		UserID:  NewULID(),
	}
	err = db.Put(email)
	assert.NoError(t, err)

	text := "Test123!"
	pass := &Password{UserID: email.UserID}
	err = pass.Generate(text)
	assert.NoError(t, err)

	err = pass.Compare(text)
	assert.NoError(t, err)

	err = db.Put(pass)
	assert.NoError(t, err)

	err = db.Put(&User{ID: email.UserID})
	assert.NoError(t, err)

	basic := []byte(fmt.Sprintf("%s:%s", email.Address, text))

	var c *Credentials
	c, err = ParseCredentials(base64.StdEncoding.EncodeToString(basic))
	assert.NoError(t, err)
	assert.Equal(t, email.Address, c.Username)
	assert.Equal(t, text, c.Password)

	err = c.ValidateUsername()
	assert.NoError(t, err)
	err = c.ValidatePassword()
	assert.NoError(t, err)

	var tkn string
	tkn, err = NewJWT(email.UserID)
	assert.NoError(t, err)
	t.Log(tkn)

	var userID ulid.ULID
	userID, err = ParseJWT(tkn)
	assert.NoError(t, err)
	assert.Equal(t, email.UserID.String(), userID.String())

	userID, err = c.Authenticate()
	assert.NoError(t, err)
	assert.Equal(t, email.UserID.String(), userID.String())
}
