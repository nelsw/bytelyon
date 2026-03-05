package test

import (
	"testing"
	"time"

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

func Test_Token(t *testing.T) {

	//db.Drop(tokenTable())
	db.Create(&model.Token{})

	var err error

	var exp = model.NewToken(ulid.Make(), model.ConfirmEmailTokenType)

	assert.NoError(t, db.Put(exp))

	var act model.Token
	act, err = db.Get[model.Token](&model.Token{ID: exp.ID})
	assert.NoError(t, err)
	assert.Equal(t, exp.Type, model.ConfirmEmailTokenType)
	assert.Equal(t, exp.UserID.String(), act.UserID.String())
	assert.Equal(t, exp.ID.String(), act.ID.String())
	assert.True(t, exp.Expiry.Add(25*time.Minute).After(time.Now().UTC()))

	err = db.Delete(exp)
	assert.NoError(t, err)

	act, err = db.Get[model.Token](&model.Token{ID: exp.ID})
	assert.Empty(t, act)
	assert.NotNil(t, act)
	assert.NoError(t, err)
	assert.True(t, act.UserID.IsZero())
}
