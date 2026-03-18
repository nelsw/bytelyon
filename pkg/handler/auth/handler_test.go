package auth

import (
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/nelsw/bytelyon/internal/config"
	"github.com/nelsw/bytelyon/internal/logger"
	"github.com/nelsw/bytelyon/pkg/db"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"
)

func init() {
	config.Init()
	logger.Init()
}

func TestHandler(t *testing.T) {

	config.Init()
	logger.Init()
	t.Setenv("JWT_SECRET", "070bb74c675267dc15a1f9466b115e57348326a30511d748712835745c5b64a8")

	var err error

	email := model.Email{
		Address: gofakeit.Email(),
		UserID:  model.NewULID(),
	}
	err = db.Put(email)
	assert.NoError(t, err)

	pass := model.Password{UserID: email.UserID}
	err = pass.Generate("Test123!")
	assert.NoError(t, err)

	err = db.Put(pass)
	assert.NoError(t, err)

	err = db.Put(&model.User{ID: email.UserID})
	assert.NoError(t, err)

	var tkn string

	tkn, err = model.NewJWT(email.UserID)
	assert.NoError(t, err)

	var userID ulid.ULID
	userID, err = model.ParseJWT(tkn)
	assert.NoError(t, err)

	assert.Equal(t, email.UserID.String(), userID.String())
}
