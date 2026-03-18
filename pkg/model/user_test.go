package model

import (
	"testing"

	"github.com/nelsw/bytelyon/pkg/db"
	"github.com/stretchr/testify/assert"
)

func Test_User(t *testing.T) {
	t.Setenv("MODE", "release")
	var err error
	exp := NewUser()
	act := new(User)

	// migrate
	db.Migrate(&User{})

	// put
	assert.NoError(t, db.Put(exp))
	assert.NoError(t, db.Put(NewUser()))
	assert.NoError(t, db.Put(NewUser()))

	// get
	act, err = db.Get(&User{ID: exp.ID})
	assert.NoError(t, err)
	assert.Equal(t, exp.ID, act.ID)

	// scan
	var arr []*User
	arr, err = db.Scan(&User{})
	size := len(arr)
	assert.NoError(t, err)
	assert.NotEmpty(t, size)

	// delete
	assert.NoError(t, db.Delete(arr[0]))
	arr, err = db.Scan(&User{})
	assert.NoError(t, err)
	assert.Len(t, arr, size-1)
}
