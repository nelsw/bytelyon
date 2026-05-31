package user

import (
	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
)

type Model struct {
	ID  ulid.ULID `json:"-"`
	EID uuid.UUID `json:"eid"`
	PID uuid.UUID `json:"pid"`
}
