package id

import (
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
)

func NewUUID(url string) uuid.UUID {
	return uuid.NewSHA1(uuid.NameSpaceURL, []byte(url))
}

func NewULID(args ...time.Time) ulid.ULID {

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

func ParseULID(id string) ulid.ULID {
	ID, err := ulid.Parse(id)
	if err != nil {
		return ulid.Zero
	}
	return ID
}
