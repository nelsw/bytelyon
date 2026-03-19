package model

import (
	"testing"
	"time"

	"github.com/nelsw/bytelyon/pkg/db"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"
)

func Test_Token(t *testing.T) {

	var err error

	var exp = NewToken(ulid.Make(), ConfirmEmailTokenType)

	//assert.NoError(t, db.Put(exp))

	var act *Token
	act, err = db.Get(&Token{ID: exp.ID})
	assert.NoError(t, err)
	assert.Equal(t, exp.Type, ConfirmEmailTokenType)
	assert.Equal(t, exp.UserID.String(), act.UserID.String())
	assert.Equal(t, exp.ID.String(), act.ID.String())
	assert.True(t, exp.Expiry.Add(25*time.Minute).After(time.Now().UTC()))

	err = db.Delete(exp)
	assert.NoError(t, err)

	act, err = db.Get(&Token{ID: exp.ID})
	assert.Empty(t, act)
	assert.NotNil(t, act)
	assert.NoError(t, err)
	assert.True(t, act.UserID.IsZero())
}
