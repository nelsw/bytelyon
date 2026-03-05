package test

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/nelsw/bytelyon/internal/logger"
	"github.com/nelsw/bytelyon/pkg/db"
	"github.com/nelsw/bytelyon/pkg/model"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"
)

func init() {
	//config.Init()
	logger.Init()
}

func Test_Email(t *testing.T) {

	db.Drop(model.emailTable())
	db.Create(model.Email{})

	var err error
	var exp = model.Email{
		Address:   gofakeit.Email(),
		UserID:    ulid.Make(),
		CreatedAt: time.Now().UTC(),
	}
	assert.NoError(t, db.Put(&exp))

	var act model.Email
	act, err = db.Get[model.Email](&model.Email{Address: exp.Address})
	assert.NoError(t, err)
	assert.Equal(t, exp.Address, act.Address)
	assert.Equal(t, exp.CreatedAt, act.CreatedAt)
	assert.True(t, act.VerifiedAt.IsZero())

	var arr []model.Email
	arr, err = db.Scan[model.Email](&model.Email{})
	assert.NoError(t, err)
	assert.NotEmpty(t, arr)

	err = db.Delete(&exp)
	assert.NoError(t, err)

	act, err = db.Get[model.Email](&model.Email{Address: exp.Address})
	assert.Empty(t, act)
	assert.NotNil(t, act)
	assert.NoError(t, err)
	assert.True(t, act.UserID.IsZero())
}
