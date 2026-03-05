package model

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/nelsw/bytelyon/internal/logger"
	"github.com/nelsw/bytelyon/pkg/db"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"
)

func init() {
	//config.Init()
	logger.Init()
}

func Test_Email(t *testing.T) {

	db.Drop(emailTable())
	db.Create(Email{})

	var err error
	var exp = Email{
		Address:   gofakeit.Email(),
		UserID:    ulid.Make(),
		CreatedAt: time.Now().UTC(),
	}
	assert.NoError(t, db.Put(&exp))

	var act Email
	act, err = db.Get[Email](&Email{Address: exp.Address})
	assert.NoError(t, err)
	assert.Equal(t, exp.Address, act.Address)
	assert.Equal(t, exp.CreatedAt, act.CreatedAt)
	assert.True(t, act.VerifiedAt.IsZero())

	var arr []Email
	arr, err = db.Scan[Email](&Email{})
	assert.NoError(t, err)
	assert.NotEmpty(t, arr)

	err = db.Delete(&exp)
	assert.NoError(t, err)

	act, err = db.Get[Email](&Email{Address: exp.Address})
	assert.Empty(t, act)
	assert.NotNil(t, act)
	assert.NoError(t, err)
	assert.True(t, act.UserID.IsZero())
}
