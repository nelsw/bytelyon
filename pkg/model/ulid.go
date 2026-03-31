package model

import (
	"math/rand"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

func NewULID(args ...time.Time) ulid.ULID {

	var t = time.Now().UTC()
	if len(args) > 0 {
		t = args[0]
	}

	id, err := ulid.New(
		ulid.Timestamp(t),
		rand.New(rand.NewSource(time.Now().UnixNano())),
	)

	if err != nil {
		id = ulid.Make()
	}

	return id
}

func ParseULID(s string) ulid.ULID {
	id, err := ulid.ParseStrict(s)
	if err == nil {
		return id
	}
	log.Warn().Str("id", s).Msg("failed to parse ULID")
	return NewULID()
}
