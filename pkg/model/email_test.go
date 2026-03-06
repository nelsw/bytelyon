package model

import (
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/nelsw/bytelyon/pkg/db"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"
)

func Test_Email(t *testing.T) {

	//db.Drop(emailTable())
	db.Create(Email{})

	var err error
	var exp = Email{
		Address: gofakeit.Email(),
		UserID:  ulid.Make(),
	}
	assert.NoError(t, db.Put(&exp))

	var act Email
	act, err = db.Get[Email](&Email{Address: exp.Address})
	assert.NoError(t, err)
	assert.Equal(t, exp.Address, act.Address)
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
