package model

import (
	"github.com/oklog/ulid/v2"
)

type User struct{ Model }

func NewUser(id ulid.ULID) *User {
	return &User{Model{UserID: id}}
}

func (u *User) ID() ulid.ULID {
	return u.Model.UserID
}
