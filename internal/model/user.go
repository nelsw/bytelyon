package model

import "github.com/google/uuid"

type User struct{ Model }

func NewUser(id uuid.UUID) *User {
	return &User{Model{UserID: id}}
}

func (u *User) ID() uuid.UUID {
	return u.Model.UserID
}
