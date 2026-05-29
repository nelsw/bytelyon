package id

import (
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/nelsw/bytelyon/pkg/util"
	"github.com/oklog/ulid/v2"
)

func NewUUID(s ...string) uuid.UUID {
	if len(s) > 0 {
		return uuid.NewSHA1(uuid.NameSpaceURL, []byte(s[0]))
	}
	return util.Safe(uuid.NewV7())
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
