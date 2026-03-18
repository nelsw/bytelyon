package model

import (
	"testing"
	"time"

	"github.com/nelsw/bytelyon/pkg/db"
	"github.com/stretchr/testify/assert"
)

func TestBot_News(t *testing.T) {

	t.Setenv("MODE", "release")

	exp := Bot{
		Type:      NewsBotType,
		UserID:    NewULID(),
		Target:    "btc price today",
		Frequency: time.Hour,
		ID:        NewULID(),
		WorkedAt:  time.Now().Add(time.Hour * 24 * 365 * 10 * -1).UTC(),
		BlackList: []string{"publix"},
	}

	act := Bot{
		UserID: exp.UserID,
		Type:   NewsBotType,
	}

	// put
	assert.NoError(t, db.PutItem(&exp))
	assert.NoError(t, db.PutItem(&Bot{
		Type:      NewsBotType,
		UserID:    exp.UserID,
		Target:    "eth price today",
		Frequency: time.Minute * 30,
		ID:        NewULID(),
	}))
	assert.NoError(t, db.PutItem(&Bot{
		Type:      NewsBotType,
		UserID:    NewULID(),
		Target:    "btc price today",
		Frequency: time.Hour * 24,
		ID:        NewULID(),
	}))

	// query
	out, err := db.Query(&act)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(out))
	for _, v := range out {
		t.Log(v.String())
	}
	// get
	act.Target = exp.Target
	_, err = db.Get(&act)
	assert.NoError(t, err)
	assert.Equal(t, exp.UserID, act.UserID)
	assert.Equal(t, exp.Target, act.Target)
	assert.Equal(t, exp.ID, act.ID)
	assert.Equal(t, exp.Frequency, act.Frequency)
	assert.Equal(t, exp.BlackList, act.BlackList)
	t.Log(act.String())
}
