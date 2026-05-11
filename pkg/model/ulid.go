package model

import (
	"math/rand"
	"time"

	"github.com/oklog/ulid/v2"
)

func NewULID(args ...time.Time) ulid.ULID {

	var t time.Time
	if len(args) > 0 {
		t = args[0]
	} else {
		t = time.Now()
	}

	id, err := ulid.New(
		ulid.Timestamp(t.UTC()),
		rand.New(rand.NewSource(time.Now().UnixNano())),
	)

	if err != nil {
		id = ulid.Make()
	}

	return id
}
