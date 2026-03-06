package model

import (
	"math/rand"
	"time"

	"github.com/oklog/ulid/v2"
)

var NewULID = func() ulid.ULID {
	id, err := ulid.New(
		ulid.Timestamp(time.Now()),
		rand.New(rand.NewSource(time.Now().UnixNano())),
	)
	if err != nil {
		id = ulid.Make()
	}
	return id
}
