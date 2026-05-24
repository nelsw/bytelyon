package id

import (
	"math/rand"
	"time"

	"github.com/oklog/ulid/v2"
)

var Descending = func(a ulid.ULID, z ulid.ULID) int {
	return z.Compare(a)
}

func New(args ...time.Time) ulid.ULID {

	var t time.Time
	if len(args) > 0 {
		t = args[0]
	}

	if t.IsZero() {
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
