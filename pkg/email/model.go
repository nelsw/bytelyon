package email

import (
	"github.com/oklog/ulid/v2"
)

type Model struct {
	UID ulid.ULID `json:"uid"`
	Txt string    `json:"txt"`
}
