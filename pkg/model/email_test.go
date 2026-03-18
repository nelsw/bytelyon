package model

import (
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/nelsw/bytelyon/pkg/db"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"
)

func Test_Email(t *testing.T) {
	t.Setenv("MODE", "release")

	var err error
	var exp = Email{
		Address: gofakeit.Email(),
		UserID:  ulid.Make(),
	}
	assert.NoError(t, db.Put(&exp))

	var act *Email
	act, err = db.Get(&Email{Address: exp.Address})
	assert.NoError(t, err)
	assert.Equal(t, exp.Address, act.Address)
	assert.Equal(t, exp.UserID, act.UserID)
	assert.True(t, act.VerifiedAt.IsZero())

	var arr []*Email
	arr, err = db.Scan(&Email{})
	assert.NoError(t, err)
	assert.NotEmpty(t, arr)

	err = db.Delete(&exp)
	assert.NoError(t, err)

	arr, err = db.Scan(&Email{})
	assert.NoError(t, err)
	assert.Empty(t, arr)
}
